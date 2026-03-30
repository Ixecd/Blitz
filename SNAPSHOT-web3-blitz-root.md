# SNAPSHOT — web3-blitz

> 项目整体快照，新会话开始时直接扔给 Claude，5 秒对齐，继续工作。
> 最后更新：2026-03-30 / v0.1.12

---

## 项目是什么

BTC/ETH 充提币系统，同时作为 dtk 的活体验证环境。

**仓库**：github.com/Ixecd/web3-blitz
**部署**：`dtk deploy`，namespace: web3-blitz

---

## 当前 K8s 状态

```
namespace: web3-blitz
  dev-toolkit-controller（Deployment）  1/1 Running ✅
  wallet-service（Deployment ×2）       1/1 Running ✅  v0.1.12
  web3-blitz-etcd-0（StatefulSet）      1/1 Running ✅  PVC
  web3-blitz-postgres-0（StatefulSet）  1/1 Running ✅  PVC
```

---

## 代码结构

```
web3-blitz/
├── cmd/wallet-service/
├── internal/
│   ├── api/
│   │   ├── handler.go    # Handler struct + NewHandler
│   │   ├── auth.go       # Register/Login/Refresh/Logout/GetMe/ForgotPassword/ResetPassword
│   │   ├── wallet.go     # GenerateAddress/GetBalance/ListDeposits/Withdraw/ListWithdrawals
│   │   ├── admin.go      # ListUsers/UpgradeUser/ListWithdrawalLimits/UpdateWithdrawalLimit
│   │   └── mux.go
│   ├── db/migrations/    # golang-migrate，启动自动执行
│   ├── email/            # SMTP（QQ 邮箱）
│   ├── lock/             # 分布式锁
│   └── wallet/btc/ eth/  # HD 钱包、Watcher
├── deployments/web3-blitz/
│   ├── web3-blitz-postgres/   StatefulSet + PVC
│   ├── web3-blitz-etcd/       StatefulSet + PVC
│   ├── wallet-service/        secretKeyRef 注入敏感变量
│   └── web3-blitz-controller/ 最小权限 RBAC
├── configs/project.env components.yaml resources.yaml
└── scripts/create-secret.sh   幂等创建 K8s Secret
```

---

## 部署流程

```bash
./scripts/create-secret.sh   # 首次或密码变更时
dtk deploy
```

---

## Secret 管理

| 变量 | 存储 |
|------|------|
| DATABASE_URL | K8s Secret（K8s service 名） |
| WALLET_HD_SEED | K8s Secret |
| SMTP_PASS | K8s Secret |
| JWT_SECRET | K8s Secret |
| ETCD_ENDPOINTS 等 | values.yaml |

---

## API 概览

```
POST /api/v1/register / login / refresh / logout
GET  /api/v1/me
POST /api/v1/forgot-password / reset-password
POST /api/v1/address          生成充值地址（BTC/ETH）
GET  /api/v1/balance / deposits / withdrawals
POST /api/v1/withdraw
GET  /api/v1/admin/users / limits
POST /api/v1/admin/users/:id/upgrade
PUT  /api/v1/admin/limits
```

---

## 接下来（P0）

1. 充值流程联调（GenerateAddress → Deposit Watcher → 确认入账）
2. 提币流程联调（余额校验 → 广播 → 状态追踪）
3. Dashboard 接真实数据
4. Prometheus + Grafana 监控配置

---

## 已知问题

| # | 问题 | 优先级 |
|---|------|--------|
| 1 | etcd 旧 registry 数据需要清理 | P1 |
| 2 | ETH geth 403（geth 配置问题） | P1 |
| 3 | FRONTEND_URL 写死 localhost，上主网前改 | P1 |
| 4 | dtk down 会删 Secret，重部署前需重建 | 已文档化 |
