# web3-blitz 项目快照

> 用途：新会话开始时直接把这个文件扔给 Claude，5秒对齐，继续工作。
> 最后更新：2026-03-18

---

## 项目是什么

Go + 云原生的交易所钱包充提币系统。
脚手架工具：`github.com/Ixecd/dev-toolkit`（dtk），支持 1 行 init boilerplate + 1 键 AI-plan + Helm deploy。

---

## 当前版本：v0.1.1

### 已完成

- BTC HD 钱包地址派生（BIP44，go-bip32）
- ETH HD 钱包地址派生（BIP44）
- BTC Deposit Watcher（每3s扫块，regtest）
- ETH Deposit Watcher（每5s扫块，geth dev）
- AddressRegistry（内存映射 address→userID，读写锁，启动从DB恢复）
- PostgreSQL + sqlc 持久化（deposit_addresses / deposits / withdrawals 三张表）
- REST API（internal/api 包，Handler + NewMux）
- BTC 提币（SendToAddress，委托 bitcoind 热钱包）
- ETH 提币（手动构建交易 + EIP-155 签名 + 广播，热钱包私钥从环境变量注入）
- 余额校验（已确认充值 - 已完成提币 = 可用余额，提币前拦截）
- 确认数逻辑（BTC 6块，ETH 12块，ConfirmChecker 每30s轮询）
- 提币历史查询（GET /api/v1/withdrawals，干净 DTO）
- Prometheus metrics
- K8s 一键部署（dtk deploy + Helm）

### API 列表

| Method | Path                  | 说明                                     |
| ------ | --------------------- | ---------------------------------------- |
| POST   | /api/v1/address       | 生成充值地址                             |
| GET    | /api/v1/balance       | 查询链上余额                             |
| GET    | /api/v1/deposits      | 查询充值记录（by user_id）               |
| GET    | /api/v1/balance/total | 查询累计已确认充值（by user_id + chain） |
| POST   | /api/v1/withdraw      | 发起提币（含余额校验）                   |
| GET    | /api/v1/withdrawals   | 查询提币历史（by user_id）               |
| GET    | /metrics              | Prometheus                               |

---

## 目录结构（关键部分）

```
internal/
├── api/
│   ├── handler.go       # 所有 HTTP handler（含 Withdraw + ListWithdrawals）
│   └── server.go        # NewMux，路由注册
├── db/
│   ├── schema.sql       # PostgreSQL 表结构（BIGSERIAL / NUMERIC(20,8) / TIMESTAMPTZ）
│   ├── queries.sql      # @param 风格，sqlc generate
│   └── queries.sql.go   # sqlc 生成
├── wallet/
│   ├── btc/
│   │   ├── btc.go       # BTCWallet（GenerateDepositAddress, GetBalance）
│   │   ├── withdraw.go  # BTC 提币（SendToAddress）
│   │   └── watcher.go   # BTC DepositWatcher（confirmed=false写入）
│   ├── eth/
│   │   ├── eth.go       # ETHWallet（GenerateDepositAddress, GetBalance）
│   │   ├── withdraw.go  # ETH 提币（EIP-155 签名广播）
│   │   └── watcher.go   # ETH DepositWatcher（confirmed=false写入）
│   ├── core/
│   │   ├── hd.go        # HDWallet（BIP44 派生）
│   │   └── confirm.go   # ConfirmChecker（BTC 6块，ETH 12块）
│   └── types/           # 共享类型（Chain, AddressRegistry, DepositRecord等）
cmd/
└── wallet-service/main.go
scripts/
└── test_withdraw.sh      # 9步 BTC 充提币 e2e 测试脚本（USER 用时间戳）
test/
└── e2e/
    └── withdraw_test.go  # Go e2e 测试（Testify）
docs/
└── guide/zh-CN/
    └── withdraw.md       # 提币流程完整文档
```

---

## 本地启动

```bash
# 启动所有依赖（bitcoind + geth + postgres）
docker compose up -d

# BTC 初始化（第一次）
bitcoin-cli -regtest createwallet "blitz_wallet"
bitcoin-cli -regtest -rpcwallet=blitz_wallet -generate 101

# 启动服务
DATABASE_URL=postgres://blitz:blitz@localhost:5432/blitz?sslmode=disable \
ETH_HOT_WALLET_KEY=<私钥hex> \
go run cmd/wallet-service/main.go
```

ETH 热钱包私钥获取（geth dev 模式）：

```bash
docker cp <geth容器ID>:"$(docker exec <容器ID> find / -name 'UTC--*' 2>/dev/null | head -1)" ./keystore.json
python3 -m venv /tmp/v && source /tmp/v/bin/activate && pip install eth-account
python3 -c "
from eth_account import Account; import json
ks = json.load(open('./keystore.json'))
acc = Account.from_key(Account.decrypt(ks, ''))
print(acc.key.hex())
"
```

---

## 测试

```bash
# BTC 充提币完整流程（含确认数校验）
./scripts/test_withdraw.sh

# Go e2e 测试（需服务已启动）
go test ./test/e2e/... -v -timeout 120s
```

---

## 待实现（按优先级）

### 1. Redis 防重复提币

同一笔提币请求不能广播两次，用 Redis 分布式锁实现。
位置：`handler.go` Withdraw，`CreateWithdrawal` 之前加锁，广播完释放。
锁 key：`withdraw:lock:{user_id}:{chain}`

### 2. BTC fee 回填

`SendToAddress` 不直接返回 fee，需通过 `GetTransaction(txid)` 回查后更新 DB。
位置：`btc/withdraw.go`，广播成功后异步查询。

### 3. 多签支持

### 4. CI/CD（GitHub Actions）

---

## 技术决策备忘

- HD 钱包底层库：`github.com/tyler-smith/go-bip32`（替换了有历史包袱的 btcutil/hdkeychain）
- BTC 地址：P2WPKH bech32，网络参数 RegressionNetParams（上主网改 MainNetParams）
- ETH 提币签名：EIP-155（含 chainID，防跨链重放）
- Nonce：从 PendingNonceAt 获取，防止 ETH 交易重放
- 提币策略：先落库（pending）→ 广播 → 更新 status（completed/failed），保证可追溯
- 余额校验：`confirmed充值总额 - completed提币总额 = 可用余额`，两个 SQL 查询相减
- 确认数：watcher 写入 confirmed=0，ConfirmChecker 每30s轮询，块高差达标后更新 confirmed=1
- ETH 热钱包私钥：环境变量注入，K8s 生产走 Secret
- DB：PostgreSQL（NUMERIC(20,8) 存金额，避免浮点精度问题），sqlc 生成 string 类型，handler 层转 float64 返回
- BTC fee 当前存 0（已知 TODO，需 GetTransaction 回填）
- test_withdraw.sh 用时间戳用户（`USER="test-$(date +%s)"`）避免历史数据干扰

---

## 环境变量

| 变量               | 说明               | 默认值                                                      |
| ------------------ | ------------------ | ----------------------------------------------------------- |
| DATABASE_URL       | PostgreSQL 连接串  | postgres://blitz:blitz@localhost:5432/blitz?sslmode=disable |
| ETH_HOT_WALLET_KEY | ETH 热钱包私钥 hex | 空（提币不可用）                                            |
| WALLET_HD_SEED     | HD 钱包种子        | 空（使用测试 seed）                                         |
| BTC_RPC_HOST       | bitcoind RPC 地址  | localhost:18443/wallet/blitz_wallet                         |
| ETH_RPC_HOST       | geth RPC 地址      | http://localhost:8545                                       |

---

## 常见问题

**BTC 提币报 insufficient funds**：blitz_wallet 余额不足，执行 `-generate 101` 补充。

**ETH 提币报"热钱包未配置"**：启动时未传 `ETH_HOT_WALLET_KEY`。

**确认后余额还是 0**：挖块数不够，BTC 需要 6 块，脚本用 `-generate 6`。

**Step 7 没有被拦截**：USER 用了固定值，改成 `USER="test-$(date +%s)"`。

**geth CrashLoopBackOff（K8s）**：deployment 加 `enableServiceLinks: false`。

**postgres 首次启动建表失败**：确认 `schema.sql` 挂载路径正确，或手动执行：

```bash
docker exec -i postgres psql -U blitz -d blitz < internal/db/schema.sql
```