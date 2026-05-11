# SNAPSHOT — blitz

**里程碑**：A2 Reconciliation Controller 完整验证通过
**日期**：2026-03-25
**版本**：v0.1.10

---

## 本次完成

### A2 Reconciliation Controller 部署并验证

Controller 作为独立 Deployment 运行在 K8s 里，监控 wallet-service 等关键资源，
检测到缺失时自动执行 `helm rollback` 恢复，全程无人工干预。

**验证过程**：
```
1. dtk deploy → 部署成功，revision 25，wallet-service v0.1.10 Running
2. kubectl delete deployment wallet-service → 手动删除
3. controller 检测到缺失（8s 周期触发）
4. helm rollback blitz revision=25
5. wallet-service 自动恢复，1/1 Running，v0.1.10 镜像正确
```

**controller 日志**：
```
资源缺失，启动自愈 kind=Deployment name=wallet-service namespace=blitz
执行 helm rollback 恢复资源 release=blitz revision=25
自愈成功 release=blitz
```

---

## 新增文件清单

### Helm chart 新增

```
deployments/blitz/templates/
├── controller-deployment.yaml   # controller Deployment，serviceAccountName 绑定
├── controller-rbac.yaml         # ServiceAccount + Role（全权限）+ RoleBinding
└── resources-configmap.yaml     # resources.yaml 挂进 controller pod
```

### 配置新增

```
configs/
└── resources.yaml               # controller 监控的资源配置（本地开发用）
```

### Dockerfile 新增

```
build/docker/blitz-controller/
└── Dockerfile                   # 基于 blitz 二进制，内含 kubectl + helm
```

---

## 关键配置说明

### values.yaml 新增字段

```yaml
global:
  projectName: blitz
  arch: arm64
  version: v0.1.10          # 统一版本号，wallet-service 和 controller 都用这个

controller:
  image:
    repository: qingchun22/blitz-controller
  etcdEndpoints: "etcd:2379"
```

### resources.yaml

```yaml
resources:
  - kind: Deployment
    name: wallet-service
    namespace: blitz
    on_missing: recreate     # 缺失时触发自愈
    max_retry: 3
    fallback: rollback

  - kind: StatefulSet
    name: postgres
    namespace: blitz
    on_missing: alert        # 只告警，不自动处理
    max_retry: 0
    fallback: ""
```

### RBAC 权限（Role）

controller 执行 helm rollback 需要完整权限，当前配置：
```yaml
rules:
  - apiGroups: [""]
    resources: ["*"]
    verbs: ["*"]
  - apiGroups: ["apps"]
    resources: ["*"]
    verbs: ["*"]
  - apiGroups: ["rbac.authorization.k8s.io"]
    resources: ["*"]
    verbs: ["*"]
  - apiGroups: ["batch"]
    resources: ["*"]
    verbs: ["*"]
  - apiGroups: ["networking.k8s.io"]
    resources: ["*"]
    verbs: ["*"]
```

> ⚠️ 生产环境需要收紧权限，当前为了打通流程使用了宽松配置。

---

## 踩过的坑（血泪史）

### 1. controller 镜像需要在 blitz 目录构建

controller 镜像包含 dtk 二进制，必须在 blitz 目录构建：

```bash
cd ~/blitz
docker build --no-cache -f build/docker/controller/Dockerfile \
  -t qingchun22/blitz-controller-arm64:<version> .
docker push qingchun22/blitz-controller-arm64:<version>
```

⚠️ 注意：必须加 `--no-cache`，否则 COPY . . 层会用缓存，代码改动不生效。

### 2. helm rollback 需要大量 RBAC 权限

helm rollback 会验证 release 里所有资源的权限，包括：
- secrets（存储 release 历史）
- serviceaccounts
- roles / rolebindings
- deployments / statefulsets

权限不够时会报 `is forbidden`，一条一条加会陷入循环，直接给全权限最省事。

### 3. wallet-service 镜像版本问题

原来 `wallet-service-deployment.yaml` 里镜像是 `v0.1.0` 硬编码，
`dtk deploy` 用 `kubectl set image` 更新镜像但 helm release 不知道，
导致 rollback 时总是回到 `v0.1.0`。

**修复**：改用 `global.version` 统一管理版本：
```yaml
image: "{{ .Values.walletService.image.repository }}:{{ .Values.global.version }}"
```

`deploy.mk` 也要传入：
```makefile
--set global.version=$(VERSION) \
--set global.arch=$(ARCH) \
```

### 4. helm pending-rollback 死锁

controller 疯狂 rollback + dtk deploy 并发 → helm 进入 pending-rollback 死锁。

**解法**：
```bash
# 停掉 controller
kubectl scale deployment/blitz-controller -n blitz --replicas=0

# 清理 pending secrets
kubectl delete secret -n blitz \
  $(kubectl get secret -n blitz -l owner=helm,name=blitz \
    -o jsonpath='{.items[?(@.metadata.labels.status=="pending-rollback")].metadata.name}')
```

### 5. `--force-conflicts` 和 `--server-side` 必须同时使用

`deploy.mk` 里加了 `--force-conflicts` 但没加 `--server-side`，helm 报错：
```
invalid client update option(s): forceConflicts enabled when serverSideApply disabled
```

直接删掉 `--force-conflicts` 解决。

### 6. 镜像名 `-arm64-arm64` 重复

`controller-deployment.yaml` 里写了：
```yaml
image: "{{ .Values.controller.image.repository }}-{{ .Values.global.arch }}:..."
```

`repository` 已经包含 arch（`qingchun22/blitz-controller`），不需要再拼，
但 wallet-service 的 repository 是 `qingchun22/wallet-service-arm64`，已经含 arch，
所以 wallet-service 不拼 arch，controller 拼 arch，两者处理方式不同。

### 7. docker build 缓存问题

blitz 代码改了但镜像没更新，因为 docker build cache 命中了 `COPY . .`。
必须用 `--no-cache` 强制重新构建。

### 8. 状态机和 controller 状态不同步

controller 触发 helm rollback 改变了 K8s 状态，但没有更新 dtk 的状态文件，
导致 `dtk deploy` 报 `当前部署状态为 ROLLING_BACK，不能发起新部署`。

临时修复：手动重置状态文件：
```bash
python3 -c "
import json
with open('/Users/qc/.dtk/state/blitz/blitz.json') as f:
    d = json.load(f)
d['state'] = 'RUNNING'
with open('/Users/qc/.dtk/state/blitz/blitz.json', 'w') as f:
    json.dump(d, f, indent=2, ensure_ascii=False)
"
```

根本解法：controller 执行 rollback 时应该同步更新 etcd 状态。

---

## 遗留 Bug / TODO

| # | 问题 | 优先级 |
|---|------|--------|
| 1 | controller rollback 后状态机不同步，dtk deploy 会报 ROLLING_BACK | P0 |
| 2 | controller RBAC 权限过宽，生产环境需要收紧 | P1 |
| 3 | controller 镜像需要手动构建推送，没有和 dtk release 联动 | P1 |
| 4 | helm pending-rollback 死锁需要手动清理 | P1 |
| 5 | etcd Watch 第一次连接会有 warn（etcd 还没就绪），无影响 | P2 |

---

## 历史快照

```
snapshots/
├── SNAPSHOT-blitz-2026-03-23-auth-migrate.md
├── SNAPSHOT-blitz-2026-03-23-k8s-deploy.md
├── SNAPSHOT-blitz-2026-03-24-fully-deployed.md
├── SNAPSHOT-blitz-2026-03-25-deposit-admin.md
└── SNAPSHOT-blitz-2026-03-25-controller.md   ← 本次
```
