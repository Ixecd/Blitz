# web3-blitz 项目快照

> 用途：新会话开始时直接把这个文件扔给 Claude，5秒对齐，继续工作。
> 最后更新：2026-03-19

---

## 项目是什么

Go + 云原生的交易所钱包充提币系统。
脚手架工具：`github.com/Ixecd/dev-toolkit`（dtk），支持 1 行 init boilerplate + 1 键 AI-plan + Helm deploy。

---

## 当前版本：v0.1.1

### 已完成

- BTC/ETH HD 钱包地址派生（BIP44，go-bip32）
- BTC/ETH Deposit Watcher + 确认数逻辑（BTC 6块，ETH 12块）
- AddressRegistry（内存映射，读写锁，启动从DB恢复）
- PostgreSQL + sqlc 持久化（6张表）
- BTC/ETH 提币（fee 回填 + EIP-155 签名）
- 余额校验（confirmed充值 - completed提币）
- etcd 分布式锁 + 选主 + 配置热更新
- 用户系统（注册/登录，bcrypt）
- JWT 认证（access token 1h + refresh token 7d，旋转策略）
- 每日提币限额（4等级：普通/白银/黄金/钻石，滚动24小时）
- GitHub Actions CI（build + vet + test + docker build）
- Prometheus metrics + K8s Helm 部署

### 数据库表（6张）

| 表名 | 说明 |
|------|------|
| deposit_addresses | 充值地址 |
| deposits | 充值记录（confirmed=0/1）|
| withdrawals | 提币记录（pending/completed/failed）|
| users | 用户（含 level 字段）|
| refresh_tokens | refresh token（revoked + expires_at）|
| withdrawal_limits | 等级限额配置（4条种子数据）|

### 提币限额

| 等级 | 名称 | BTC日限 | ETH日限 | 升级条件(累计BTC充值) |
|------|------|---------|---------|-------------------|
| 0 | 普通用户 | 2 | 50 | 默认 |
| 1 | 白银用户 | 10 | 200 | ≥ 1 BTC |
| 2 | 黄金用户 | 50 | 1000 | ≥ 10 BTC |
| 3 | 钻石用户 | 200 | 5000 | ≥ 50 BTC |

### API 列表

| Method | Path | 认证 | 说明 |
|--------|------|------|------|
| POST | /api/v1/register | 无 | 用户注册 |
| POST | /api/v1/login | 无 | 登录，返回 access_token + refresh_token |
| POST | /api/v1/refresh | 无 | 换新 token（旋转策略）|
| POST | /api/v1/logout | 无 | 撤销 refresh_token |
| POST | /api/v1/address | 无 | 生成充值地址 |
| GET  | /api/v1/balance | 无 | 查询链上余额 |
| GET  | /api/v1/deposits | 无 | 查询充值记录 |
| GET  | /api/v1/balance/total | 无 | 查询累计已确认充值 |
| POST | /api/v1/withdraw | **JWT** | 发起提币（余额+限额+分布式锁）|
| GET  | /api/v1/withdrawals | 无 | 查询提币历史 |
| GET  | /metrics | 无 | Prometheus |

---

## 目录结构（关键部分）

```
internal/
├── api/handler.go       # 所有 HTTP handler
├── api/server.go        # NewMux，JWT 中间件挂载
├── auth/auth.go         # HashPassword, GenerateToken, GenerateRefreshToken
├── auth/middleware.go   # JWTMiddleware, GetClaims
├── config/rpc.go        # BTCRPCHolder / ETHRPCHolder
├── config/watcher.go    # ConfigWatcher（etcd watch）
├── db/schema.sql        # 6张表结构
├── db/seed.sql          # withdrawal_limits 种子数据
├── db/queries.sql
├── lock/lock.go         # etcd 分布式锁
├── wallet/btc/          # BTC 钱包 + watcher + withdraw
├── wallet/eth/          # ETH 钱包 + watcher + withdraw
├── wallet/core/
│   ├── hd.go            # HDWallet
│   └── confirm.go       # ConfirmChecker + etcd 选主
└── wallet/types/
cmd/
├── wallet-service/main.go
├── chain-miner/main.go
└── pos-sim/main.go
.github/workflows/ci.yml
Procfile.single
```

---

## 本地启动

```bash
goreman -f Procfile.single start

docker compose up -d

DATABASE_URL=postgres://blitz:blitz@localhost:5432/blitz?sslmode=disable \
ETH_HOT_WALLET_KEY=<私钥hex> \
go run cmd/wallet-service/main.go
```

---

## etcd key 总览

| key | 用途 | TTL |
|-----|------|-----|
| `/blitz/lock/withdraw:{uid}:{chain}` | 提币分布式锁 | 30s |
| `/blitz/leader/confirm-checker` | ConfirmChecker 选主 | 15s |
| `/blitz/config/btc_rpc_host` | BTC RPC 热更新 | 永久 |
| `/blitz/config/eth_rpc_host` | ETH RPC 热更新 | 永久 |

---

## 待实现（按优先级）

### 1. user_id 统一（TEXT → BIGINT）
deposits/withdrawals.user_id 目前是 TEXT，和 users.id BIGINT 体系不同。
统一后 checkAndUpgradeLevel 才能自动升级用户等级。

### 2. ETH reorg 处理
链重组时回滚已确认充值。

### 3. consumeDeposits 死信队列
重试失败后持久化，防止充值记录丢失。

### 4. 完整 IAM
RBAC、审计日志、黑名单 token、KYC、提币限额管理接口。

### 5. 多签支持

### 6. 主网参数切换

---

## 技术决策备忘

- deposits/withdrawals.user_id 是 TEXT（历史遗留），users.id 是 BIGINT，两套体系暂时并存
- JWT claims 里的 user_id 是 BIGINT，提币限额校验用 claims.UserID
- 过去24小时提币额度查询用 withdrawals.user_id（TEXT），传 req.UserID
- refresh token 旋转：每次 /refresh 撤销旧 token，发新 token
- 种子数据在 connect.go runSeed 里执行，ON CONFLICT DO NOTHING 保证幂等
- BTC fee 回填：SendToAddress 后 GetTransaction 查实际 fee
- e2e/smoke 测试加 `//go:build e2e`，CI 只跑 integration

---

## 环境变量

| 变量 | 默认值 |
|------|--------|
| DATABASE_URL | postgres://blitz:blitz@localhost:5432/blitz?sslmode=disable |
| ETH_HOT_WALLET_KEY | 空 |
| WALLET_HD_SEED | 空（测试 seed）|
| JWT_SECRET | dev-secret |
| ETCD_ENDPOINTS | localhost:2379 |
| PORT | 2113 |
| BTC_RPC_HOST | localhost:18443/wallet/blitz_wallet |
| ETH_RPC_HOST | http://localhost:8545 |

---

## 历史快照

```
snapshots/
├── SNAPSHOT-2026-03-18-pg-migration.md
├── SNAPSHOT-2026-03-18-etcd.md
└── SNAPSHOT-2026-03-19-auth-limits.md
```
