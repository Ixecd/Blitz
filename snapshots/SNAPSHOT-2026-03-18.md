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
- BTC 提币（SendToAddress + GetTransaction fee 回填）
- ETH 提币（手动构建交易 + EIP-155 签名 + 广播）
- 余额校验（已确认充值 - 已完成提币 = 可用余额）
- 确认数逻辑（BTC 6块，ETH 12块，ConfirmChecker 每30s轮询）
- 提币历史查询（GET /api/v1/withdrawals）
- etcd 分布式锁（防重复提币，lease TTL 30s，进程崩溃自动释放）
- etcd 选主（ConfirmChecker Leader Election，lease TTL 15s，keepalive续期）
- etcd 配置热更新（BTC/ETH RPC 地址无重启切换，ConfigWatcher watch prefix）
- Prometheus metrics
- K8s 一键部署（dtk deploy + Helm）

### API 列表

| Method | Path                  | 说明                                     |
| ------ | --------------------- | ---------------------------------------- |
| POST   | /api/v1/address       | 生成充值地址                             |
| GET    | /api/v1/balance       | 查询链上余额                             |
| GET    | /api/v1/deposits      | 查询充值记录（by user_id）               |
| GET    | /api/v1/balance/total | 查询累计已确认充值（by user_id + chain） |
| POST   | /api/v1/withdraw      | 发起提币（含余额校验 + 分布式锁）        |
| GET    | /api/v1/withdrawals   | 查询提币历史（by user_id）               |
| GET    | /metrics              | Prometheus                               |

---

## 目录结构（关键部分）

```
internal/
├── api/
│   ├── handler.go       # 所有 HTTP handler
│   └── server.go        # NewMux，路由注册
├── config/
│   ├── rpc.go           # BTCRPCHolder / ETHRPCHolder（线程安全）
│   └── watcher.go       # ConfigWatcher，watch /blitz/config/ 前缀
├── db/
│   ├── schema.sql       # PostgreSQL 表结构
│   ├── queries.sql      # @param 风格
│   └── queries.sql.go   # sqlc 生成
├── lock/
│   └── lock.go          # etcd 分布式锁（lease + txn CAS）
├── wallet/
│   ├── btc/
│   │   ├── btc.go       # BTCWallet（用 BTCRPCHolder）
│   │   ├── withdraw.go  # BTC 提币 + fee 回填
│   │   └── watcher.go   # BTC DepositWatcher（用 BTCRPCHolder）
│   ├── eth/
│   │   ├── eth.go       # ETHWallet（用 ETHRPCHolder）
│   │   ├── withdraw.go  # ETH 提币（EIP-155）
│   │   └── watcher.go   # ETH DepositWatcher（用 ETHRPCHolder）
│   ├── core/
│   │   ├── hd.go        # HDWallet（BIP44 派生）
│   │   └── confirm.go   # ConfirmChecker + etcd 选主
│   └── types/           # 共享类型
cmd/
└── wallet-service/main.go
scripts/
└── test_withdraw.sh      # 9步 BTC 充提币 e2e 测试
test/
└── e2e/
    └── withdraw_test.go
docs/
└── guide/zh-CN/
    └── withdraw.md
Procfile.single           # goreman 单节点 etcd 启动
```

---

## 本地启动

```bash
# 1. 启动 etcd（单节点）
goreman -f Procfile.single start

# 2. 启动依赖
docker compose up -d   # bitcoind + geth + postgres

# 3. BTC 初始化（第一次）
bitcoin-cli -regtest createwallet "blitz_wallet"
bitcoin-cli -regtest -rpcwallet=blitz_wallet -generate 101

# 4. 启动服务
DATABASE_URL=postgres://blitz:blitz@localhost:5432/blitz?sslmode=disable \
ETH_HOT_WALLET_KEY=<私钥hex> \
go run cmd/wallet-service/main.go
```

---

## etcd 使用场景

| 场景       | key                                      | 说明                         |
| ---------- | ---------------------------------------- | ---------------------------- |
| 分布式锁   | `/blitz/lock/withdraw:{user_id}:{chain}` | 防重复提币，TTL 30s          |
| 选主       | `/blitz/leader/confirm-checker`          | ConfirmChecker 单活，TTL 15s |
| 配置热更新 | `/blitz/config/btc_rpc_host`             | BTC RPC 地址                 |
| 配置热更新 | `/blitz/config/eth_rpc_host`             | ETH RPC 地址                 |

触发热更新：

```bash
etcdctl --endpoints=localhost:2379 put \
  /blitz/config/btc_rpc_host \
  "localhost:18443/wallet/blitz_wallet"
```

---

## 测试

```bash
# 充提币完整流程
./scripts/test_withdraw.sh

# 并发重复提币测试（期望一个 429 一个 200）
curl -X POST http://localhost:2113/api/v1/withdraw \
  -H "Content-Type: application/json" \
  -d '{"user_id":"test001","to_address":"bcrt1q...","amount":0.05,"chain":"btc"}' &
curl -X POST http://localhost:2113/api/v1/withdraw \
  -H "Content-Type: application/json" \
  -d '{"user_id":"test001","to_address":"bcrt1q...","amount":0.05,"chain":"btc"}' &
wait
```

---

## 待实现

### 1. IAM（轻量 JWT 中间件）

堵住 /api/v1/withdraw 无鉴权漏洞，一天内可完成。

### 2. 多签支持

### 3. CI/CD（GitHub Actions）

---

## 技术决策备忘

- HD 钱包：`github.com/tyler-smith/go-bip32`
- BTC 地址：P2WPKH bech32，RegressionNetParams（上主网改 MainNetParams）
- ETH 提币：EIP-155 签名，PendingNonceAt 防重放
- 提币策略：先落库 pending → 广播 → 更新 completed/failed
- 余额校验：confirmed充值 - completed提币 = 可用余额
- 确认数：watcher 写 confirmed=0，ConfirmChecker 轮询更新
- DB：PostgreSQL，NUMERIC(20,8) 存金额，sqlc 生成 string，handler 层转 float64
- etcd 锁：lease + txn CAS，崩溃自动释放，无死锁风险
- etcd 选主：campaign loop + keepalive，TTL 内完成 failover
- etcd 热更新：watch prefix，重建连接后原子 swap holder
- RPC 热更新：所有组件通过 Holder.Get() 访问，swap 对调用方透明
- BTC fee：SendToAddress 后 GetTransaction 回填，regtest 约 0.00007976 BTC

---

## 环境变量

| 变量               | 说明               | 默认值                                                      |
| ------------------ | ------------------ | ----------------------------------------------------------- |
| DATABASE_URL       | PostgreSQL 连接串  | postgres://blitz:blitz@localhost:5432/blitz?sslmode=disable |
| ETH_HOT_WALLET_KEY | ETH 热钱包私钥 hex | 空（提币不可用）                                            |
| WALLET_HD_SEED     | HD 钱包种子        | 空（使用测试 seed）                                         |
| BTC_RPC_HOST       | bitcoind RPC 地址  | localhost:18443/wallet/blitz_wallet                         |
| ETH_RPC_HOST       | geth RPC 地址      | http://localhost:8545                                       |
| ETCD_ENDPOINTS     | etcd 地址          | localhost:2379                                              |
| PORT               | HTTP 服务端口      | 2113                                                        |

---

## 历史快照

```
snapshots/
└── SNAPSHOT-2026-03-18-pg-migration.md   # PostgreSQL 迁移完成时
```