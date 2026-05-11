# 死信队列设计

> blitz 充值记录写入失败的可靠性保障机制。

---

## 问题背景

DepositWatcher 扫描到新充值后，通过 channel 传递给 consumeDeposits 写入 PostgreSQL。如果 DB 写入失败（网络抖动、连接池耗尽、PostgreSQL 重启等），充值记录会永久丢失，用户的钱凭空消失。

原始实现的问题：

```go
// 写入失败直接 log，这笔充值记录永久丢失
if err := queries.CreateDeposit(ctx, params); err != nil {
    log.Printf("[ERROR] 写入失败: %v", err)
}
```

---

## 解决方案

三次重试 + 死信持久化：

```
consumeDeposits 收到充值记录
    │
    ▼ 第1次写入
    ├─ 成功 → 结束
    └─ 失败 → 等待 1s
            │
            ▼ 第2次写入
            ├─ 成功 → 结束
            └─ 失败 → 等待 2s
                    │
                    ▼ 第3次写入
                    ├─ 成功 → 结束
                    └─ 最终失败 → 写入 dead_letters 表
```

重试间隔递增（1s/2s/3s），避免在 DB 压力大时密集重试。

---

## 数据库设计

```sql
CREATE TABLE IF NOT EXISTS dead_letters (
    id         BIGSERIAL PRIMARY KEY,
    type       TEXT NOT NULL,        -- 死信类型，如 btc_deposit / eth_deposit
    payload    JSONB NOT NULL,       -- 原始参数，完整保存方便重放
    error      TEXT NOT NULL,        -- 最后一次失败的错误信息
    retries    INTEGER NOT NULL DEFAULT 0,
    resolved   BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);
```

payload 使用 JSONB 存储原始的 CreateDepositParams，方便后续重放：

```json
{
  "tx_id": "c8d0ae07...",
  "address": "bcrt1q3mc...",
  "user_id": "test001",
  "amount": "0.05000000",
  "height": 9133,
  "confirmed": 0,
  "chain": "btc"
}
```

---

## 代码实现

```go
consumeDeposits := func(ch <-chan types.DepositRecord, chainName string) {
    for deposit := range ch {
        params := db.CreateDepositParams{...}

        const maxRetries = 3
        var lastErr error
        for i := range maxRetries {
            lastErr = queries.CreateDeposit(context.Background(), params)
            if lastErr == nil {
                break
            }
            log.Printf("[WARN] 写入失败(第%d次): %v", i+1, lastErr)
            time.Sleep(time.Duration(i+1) * time.Second)
        }

        if lastErr != nil {
            payload, _ := json.Marshal(params)
            _ = queries.CreateDeadLetter(context.Background(), db.CreateDeadLetterParams{
                Type:    chainName + "_deposit",
                Payload: payload,
                Error:   lastErr.Error(),
                Retries: maxRetries,
            })
            log.Printf("[ERROR] 写入死信队列: txid=%s", deposit.TxID)
        }
    }
}
```

---

## 运维操作

查看未处理的死信：

```sql
SELECT id, type, error, retries, created_at
FROM dead_letters
WHERE resolved = FALSE
ORDER BY created_at ASC;
```

手动重放单条死信：

```sql
-- 1. 查看 payload
SELECT payload FROM dead_letters WHERE id = 1;

-- 2. 手动插入充值记录
INSERT INTO deposits (tx_id, address, user_id, amount, height, confirmed, chain)
VALUES ('c8d0ae07...', 'bcrt1q3mc...', 'test001', 0.05, 9133, 0, 'btc');

-- 3. 标记死信已处理
UPDATE dead_letters SET resolved = TRUE, updated_at = NOW() WHERE id = 1;
```

---

## 设计说明

**为什么不用 Redis 队列**：Redis 挂掉死信也会丢失，引入新组件增加运维负担。PostgreSQL 本来就在，JSONB 存 payload 足够，技术栈统一。

**为什么 payload 用 JSONB 而不是 TEXT**：JSONB 支持索引和查询，运维时可以直接 `SELECT payload->>'tx_id'` 查找特定交易。

**重试3次的依据**：1次可能是瞬时抖动，3次覆盖大多数短暂故障。超过3次说明 DB 有持续性问题，继续重试意义不大，人工介入更合适。

---

## 待完善

- 定时重试：后台 goroutine 定期扫描 resolved=FALSE 的记录，自动重放
- 管理接口：GET /api/v1/dead-letters（需要 admin 权限）
- 重放接口：POST /api/v1/dead-letters/:id/replay
- 告警：死信数量超阈值时触发 Alertmanager 通知
