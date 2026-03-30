# 项目交接文档 — web3-blitz

> 写给下一个 Claude
> 日期：2026-03-30
> 版本：v0.1.12

---

## 写在前面

web3-blitz 是 BTC/ETH 充提币系统，同时是 dtk 的活体验证环境。两个项目同步开发，web3-blitz 踩的坑直接反哺 dtk。

**qc 的工作风格**：
- 设计优先，先对齐再动手
- 喜欢被推 back，不喜欢被认同。他通常是对的
- 前端暂时不是重点，先把后端链路跑稳
- `slog` 不用 `log`，严格分包

---

## 一、当前状态

全部 Running，v0.1.12，grpc 已升级修复 CVE-2026-33186：

```
namespace: web3-blitz
  dev-toolkit-controller  1/1 Running ✅
  wallet-service (×2)     1/1 Running ✅
  web3-blitz-etcd-0       1/1 Running ✅  StatefulSet + PVC
  web3-blitz-postgres-0   1/1 Running ✅  StatefulSet + PVC
```

---

## 二、部署流程

```bash
# 首次或 dtk down 之后（Secret 会被删）
./scripts/create-secret.sh

# 部署
dtk deploy

# 安全扫描
dtk scan
```

---

## 三、接下来（P0 充提币联调）

### 充值联调
1. `POST /api/v1/address` 生成 BTC/ETH 充值地址
2. regtest 环境发送真实交易
3. 验证 Deposit Watcher 检测到入账
4. ConfirmChecker 确认数达标后写 DB

### 提币联调
1. `POST /api/v1/withdraw` 余额/限额校验
2. 广播到 regtest 节点
3. 状态追踪（pending → completed / failed）

### Dashboard
- `GET /api/v1/balance` 接真实数据
- 定时刷新

---

## 四、代码结构

```
internal/api/
├── handler.go    # Handler struct + NewHandler（只有这两个）
├── auth.go       # 认证（Register/Login/Refresh/Logout/GetMe/Forgot/Reset）
├── wallet.go     # 充提币（GenerateAddress/Balance/Deposits/Withdraw/Withdrawals）
├── admin.go      # 管理（Users/Upgrade/Limits）
└── mux.go        # 路由注册
```

---

## 五、已知问题

| # | 问题 | 优先级 |
|---|------|--------|
| 1 | etcd 旧 registry 数据需要清理 | P1 |
| 2 | ETH geth 403 Forbidden（geth 配置） | P1 |
| 3 | FRONTEND_URL 写死 localhost:5173 | P1 上主网前改 |
| 4 | dtk down 删 namespace 会连带删 Secret | 已在 gotchas 文档化 |
| 5 | bitcoind 未接入 K8s（本地 regtest） | P2 |

---

## 六、常用命令

```bash
cd ~/web3-blitz

# 部署
./scripts/create-secret.sh && dtk deploy

# 查看状态
dtk status
kubectl get pods -n web3-blitz

# 本地调试
kubectl port-forward -n web3-blitz deployment/wallet-service 2113:2113

# 安全扫描
dtk scan

# 查看日志
kubectl logs -n web3-blitz deployment/wallet-service -f
```
