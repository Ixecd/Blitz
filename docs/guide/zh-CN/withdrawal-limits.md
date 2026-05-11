# 提币限额指南

> blitz 的用户等级体系和每日提币限额设计说明。

---

## 设计理念

限额按用户等级 + 链维度管理，滚动24小时窗口计算。默认限额较高，高等级用户几乎无限制。

---

## 等级体系

| 等级 | 名称 | BTC 日限 | ETH 日限 | 升级条件（累计BTC充值）|
|------|------|---------|---------|-------------------|
| 0 | 普通用户 | 2 BTC | 50 ETH | 默认 |
| 1 | 白银用户 | 10 BTC | 200 ETH | >= 1 BTC |
| 2 | 黄金用户 | 50 BTC | 1000 ETH | >= 10 BTC |
| 3 | 钻石用户 | 200 BTC | 5000 ETH | >= 50 BTC |

---

## 限额校验流程

```
POST /api/v1/withdraw（需 JWT）
    │
    ▼
从 JWT claims 取 user_id
    │
    ▼
查询用户等级（users.level）
    │
    ▼
查询等级对应限额（withdrawal_limits）
    │
    ▼
查询过去24小时已提币总额
    │
    ▼
已用 + 本次 > 限额？
    ├─ 是 → 400 超出每日提币限额
    └─ 否 → 继续余额校验 → 广播交易
```

---

## 滚动24小时说明

不按自然日（00:00重置），而是滚动窗口：

```sql
WHERE created_at > NOW() - INTERVAL '24 hours'
```

比自然日更公平，避免用户在 23:59 和 00:01 各提一次绕过限制。

---

## 数据库设计

限额配置在 `connect.go` 的 `runSeed` 里初始化，`ON CONFLICT DO NOTHING` 保证幂等，重启服务不会重复插入。

---

## 用户等级升级

当前版本等级升级为占位符，只打日志。

原因：`deposits.user_id` 是 TEXT，`users.id` 是 BIGINT，两套体系暂时无法自动关联。

手动升级：
```sql
UPDATE users SET level = 1 WHERE username = 'alice';
```

---

## 运营配置

修改限额不需要重启，直接更新数据库：

```sql
UPDATE withdrawal_limits SET btc_daily = '20.00000000' WHERE level = 0;

SELECT level, level_name, btc_daily, eth_daily FROM withdrawal_limits ORDER BY level;
```

---

## 错误响应

```
HTTP 400
超出每日提币限额: 已用 0.15000000，本次 3.00000000，限额 2.00000000（普通用户）
```
