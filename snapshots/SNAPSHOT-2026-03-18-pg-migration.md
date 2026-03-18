# web3-blitz 项目快照

> 用途：新会话开始时直接把这个文件扔给 Claude，5秒对齐，继续工作。
> 最后更新：2026-03-18
> 历史快照：`snapshots/` 目录

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
- BTC 提币（SendToAddress + fee 回填via GetTransaction）
- ETH 提币（手动构建交易 + EIP-155 签名 + 广播）
- 余额校验（已确认充值 - 已完成提币 = 可用余额）
- 确认数逻辑（BTC 6块，ETH 12块，ConfirmChecker 每30s轮询）
- 提币历史查询（GET /api/v1/withdrawals）
- **etcd 分布式锁**（防重复提币，lease TTL 30s，进程崩溃自动释放）
- **etcd 选主**（ConfirmChecker Leader Election，lease TTL 15s，15s内完成故障转移）
- **etcd 配置热更新**（BTC/ETH RPC 地址变更无需重启，ConfigWatcher watch /blitz/config/）
- Prometheus metrics
- K8s 一键部署（dtk deploy + Helm）

### API 列表

| Method | Path | 说明 |
|--------|------|------|
| POST | /api/v1/address | 生成充值地址 |
| GET  | /api/v1/balance | 查询链上余额 |
| GET  | /api/v1/deposits | 查询充值记录（by user_id）|
| GET  | /api/v1/balance/total | 查询累计已确认充值（by user_id + chain）|
| POST | /api/v1/withdraw | 发起提币（含余额校验 + 分布式锁）|
| GET  | /api/v1/withdrawals | 查询提币历史（by user_id）|
| GET  | /metrics | Prometheus |

---

## 目录结构（关键部分）

```
internal/
├── api/
│   ├── handler.go       # 所有 HTTP handler
│   └── server.go        # NewMux，路由注册
├── config/
│   ├── rpc.go           # BTCRPCHolder / ETHRPCHolder（线程安全 RPC 持有者）
│   └── watcher.go       # ConfigWatcher（etcd watch /blitz/config/ 热更新）
├── db/
│   ├── schema.sql       # PostgreSQL 表结构（BIGSERIAL / NUMERIC(20,8) / TIMESTAMPTZ）
│   ├── queries.sql      # @param 风格，sqlc generate
│   └── queries.sql.go   # sqlc 生成
├── lock/
│   └── lock.go          # DistributedLock（etcd lease + txn CAS）
├── wallet/
│   ├── btc/
│   │   ├── btc.go       # BTCWallet（rpcHolder）
│   │   ├── withdraw.go  # BTC 提币 + queryFee 回填
│   │   └── watcher.go   # BTC DepositWatcher（rpcHolder）
│   ├── eth/
│   │   ├── eth.go       # ETHWallet（rpcHolder）
│   │   ├── withdraw.go  # ETH 提币（EIP-155）
│   │   └── watcher.go   # ETH DepositWatcher（rpcHolder）
│   ├── core/
│   │   ├── hd.go        # HDWallet（BIP44 派生）
│   │   └── confirm.go   # ConfirmChecker（etcd 选主 + BTC 6块/ETH 12块）
│   └── types/           # 共享类型
cmd/
└── wallet-service/main.go
scripts/
└── test_withdraw.sh      # 9步 BTC 充提币 e2e 测试脚本
test/
└── e2e/
    └── withdraw_test.go
docs/
└── guide/zh-CN/
    └── withdraw.md
Procfile.single           # 单节点 etcd 本地开发
```

---

## 本地启动

```bash
# 1. 启动 etcd（单节点）
goreman -f Procfile.single start

# 2. 启动依赖（bitcoind + geth + postgres）
docker compose up -d

# 3. BTC 初始化（第一次）
bitcoin-cli -regtest createwallet "blitz_wallet"
bitcoin-cli -regtest -rpcwallet=blitz_wallet -generate 101

# 4. 启动服务
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

# 并发重复提币测试（应返回 429）
curl ... & curl ... & wait

# 配置热更新测试
etcdctl --endpoints=localhost:2379 put \
  /blitz/config/btc_rpc_host \
  "localhost:18443/wallet/blitz_wallet"
```

---

## 待实现（按优先级）

### 1. 整体 code review
### 2. 多签支持
### 3. CI/CD（GitHub Actions）

---

## 技术决策备忘

- HD 钱包：`github.com/tyler-smith/go-bip32`（替换有历史包袱的 btcutil/hdkeychain）
- BTC 地址：P2WPKH bech32，RegressionNetParams（上主网改 MainNetParams）
- ETH 提币：EIP-155 签名（含 chainID 防跨链重放），Nonce 从 PendingNonceAt 获取
- 提币策略：先落库 pending → 广播 → 更新 completed/failed
- 余额校验：confirmed充值 - completed提币，两条 SQL 相减
- 确认数：watcher 写 confirmed=0，ConfirmChecker 轮询块高差达标后更新
- etcd 分布式锁：lease + txn CAS，TTL 30s，进程崩溃自动释放，无死锁风险
- etcd 选主：campaign loop，winner keepalive lease，loser 每3s重试，故障转移 ≤15s
- etcd 热更新：ConfigWatcher watch prefix，Holder 模式解耦 RPC 客户端与业务逻辑
- DB：PostgreSQL NUMERIC(20,8) 存金额避免浮点，sqlc 生成 string，handler 层转 float64
- BTC fee：GetTransaction 回查，bitcoind 返回负数取绝对值

---

## 环境变量

| 变量 | 说明 | 默认值 |
|------|------|--------|
| DATABASE_URL | PostgreSQL 连接串 | postgres://blitz:blitz@localhost:5432/blitz?sslmode=disable |
| ETH_HOT_WALLET_KEY | ETH 热钱包私钥 hex | 空（提币不可用）|
| WALLET_HD_SEED | HD 钱包种子 | 空（使用测试 seed）|
| BTC_RPC_HOST | bitcoind RPC 地址 | localhost:18443/wallet/blitz_wallet |
| ETH_RPC_HOST | geth RPC 地址 | http://localhost:8545 |
| ETCD_ENDPOINTS | etcd 地址 | localhost:2379 |
| PORT | HTTP 服务端口 | 2113 |

---

## 常见问题

**BTC 提币报 insufficient funds**：blitz_wallet 余额不足，`-generate 101` 补充。

**ETH 提币报"热钱包未配置"**：启动时未传 `ETH_HOT_WALLET_KEY`。

**确认后余额还是 0**：BTC 需要 6 块，脚本用 `-generate 6`。

**并发提币没被拦截**：检查 etcd 是否正常运行，`etcdctl endpoint health`。

**ConfirmChecker 不执行**：检查是否竞选到 Leader，日志看 `👑`。

**热更新没生效**：检查 key 是否正确 `/blitz/config/btc_rpc_host`。

**geth CrashLoopBackOff（K8s）**：deployment 加 `enableServiceLinks: false`。

**postgres 首次建表失败**：
```bash
docker exec -i postgres psql -U blitz -d blitz < internal/db/schema.sql
```
