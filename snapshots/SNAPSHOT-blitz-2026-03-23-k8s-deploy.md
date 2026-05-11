# SNAPSHOT — blitz

**里程碑**：K8s 部署调试中（postgres + embed migrations）
**日期**：2026-03-23

---

## 本次完成

### golang-migrate embed 方案
- `connect.go` 换用 `embed.FS` + `iofs`，迁移文件打进二进制
- 彻底解决容器里找不到 migrations 目录的问题
- 不需要 `MIGRATIONS_PATH` 环境变量，不需要 Dockerfile 额外 COPY

### K8s postgres
- Bitnami postgresql 已装到 `blitz` namespace
- service name：`postgres-postgresql`，port `5432`
- `values.yaml` 顶层 `env` 已加 `DATABASE_URL`

### 待完成（下次继续）
- `walletService.env` 里补上 `DATABASE_URL`（顶层 env 没生效，deployment template 用的是 walletService.env）
- 重新 `dtk deploy` 验证 pod 正常启动

---

## K8s 环境现状

```
namespace: blitz
services:
  - bitcoind        ClusterIP  18443
  - geth            ClusterIP  8545
  - postgres-postgresql  ClusterIP  5432
  - blitz      ClusterIP  80

pods:
  - bitcoind        Running ✅
  - geth-rpc        Running ✅
  - postgres-postgresql  Running ✅
  - wallet-service  CrashLoopBackOff ❌ (DATABASE_URL 未注入)
  - blitz      Running (有重启，待查)
```

---

## 下次启动步骤

```bash
# 1. values.yaml walletService.env 加 DATABASE_URL
# 2. 升版本
# 3. dtk deploy
# 4. 验证
kubectl logs -n blitz <wallet-service-pod>
```

---

## 文件变动清单

```
修改：
- internal/db/connect.go     ← embed.FS + iofs，删 runSeed，默认 DSN 加 search_path
- deployments/blitz/values.yaml  ← 顶层 env 加 DATABASE_URL（walletService.env 待加）
```

---

## 历史快照

```
snapshots/
├── SNAPSHOT-blitz-2026-03-23-auth-migrate.md
└── SNAPSHOT-blitz-2026-03-23-k8s-deploy.md   ← 本次
```
