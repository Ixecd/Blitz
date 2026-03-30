# SNAPSHOT — web3-blitz

**里程碑**：v0.1.12 安全修复 + K8s 工程化完成
**日期**：2026-03-30
**版本**：v0.1.12

---

## 本次完成

### CVE-2026-33186 修复

`google.golang.org/grpc` 从 v1.77.0 升级到 v1.79.3，修复 gRPC-Go 授权绕过漏洞（CVSS 9.1）。

由 `dtk scan` 自动扫出，修复后重扫验证：

```
[15:58:19] ✓  qingchun22/wallet-service-arm64:v0.1.12 无漏洞 ✅
```

### Makefile 同步

从 dev-toolkit 同步 `deploy.mk` 和 `tools.mk`：
- `deploy.mk`：本地 image inspect 替代远端 manifest inspect，KUBECTL/HELM flags 优化
- `tools.mk`：新增 `install.trivy` / `install.cosign`
- `Makefile`：显式指定 `BINS = wallet-service`，避免扫描到 chain-miner/pos-sim

---

## 当前 K8s 状态

```
namespace: web3-blitz
  dev-toolkit-controller（Deployment）  1/1 Running ✅
  wallet-service（Deployment ×2）       1/1 Running ✅  v0.1.12，grpc v1.79.3
  web3-blitz-etcd-0（StatefulSet）      1/1 Running ✅  PVC 持久化
  web3-blitz-postgres-0（StatefulSet）  1/1 Running ✅  PVC 持久化
```

---

## 历史快照

```
snapshots/
├── SNAPSHOT-web3-blitz-2026-03-24-fully-deployed.md
├── SNAPSHOT-web3-blitz-2026-03-25-controller.md
├── SNAPSHOT-web3-blitz-2026-03-30-k8s-hardening.md
└── SNAPSHOT-web3-blitz-2026-03-30-v0.1.12.md  ← 本次
```
