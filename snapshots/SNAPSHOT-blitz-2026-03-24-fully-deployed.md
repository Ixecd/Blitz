# SNAPSHOT — blitz

**里程碑**：K8s 全链路部署跑通 🏆
**日期**：2026-03-24

---

## 🎉 本次完成

一条 `dtk deploy` 命令，完整部署以下所有组件：

```
namespace: blitz
├── postgres-0                    1/1 Running  ✅  StatefulSet + PVC
├── etcd-8b45c4f79                1/1 Running  ✅  Deployment + emptyDir
├── bitcoind-c4f99cf88            1/1 Running  ✅
├── geth-rpc-7bd4f54664           1/1 Running  ✅
└── wallet-service-56895df778     1/1 Running  ✅  DB迁移完成，etcd选主成功
```

---

## 完整启动日志（新镜像 v0.1.4）

```json
{"level":"INFO","msg":"数据库迁移完成","version":2}
{"level":"INFO","msg":"数据库已连接"}
{"level":"INFO","msg":"etcd 已连接"}
{"level":"INFO","msg":"从 DB 恢复充值地址","count":0}
{"level":"INFO","msg":"Wallet Core 服务已启动"}
{"level":"INFO","msg":"ConfigWatcher 已启动，监听配置变更..."}
{"level":"INFO","msg":"ConfirmChecker 已启动，竞选 Leader..."}
{"level":"INFO","msg":"Deposit Watcher 已启动"}
{"level":"INFO","msg":"ETH Deposit Watcher 已启动"}
{"level":"INFO","msg":"API 服务已启动","addr":"http://localhost:2113"}
{"level":"INFO","msg":"ConfirmChecker 成为 Leader"}
```

---

## 架构演进记录

### 从手动到自动化的完整路径

**第一阶段**：手动 `docker exec psql` 建表 → 踩 pgx v5 search_path 坑，2小时

**第二阶段**：引入 golang-migrate + embed.FS → 迁移文件打进二进制，search_path 问题彻底消灭

**第三阶段**：Bitnami Helm dependency → 遭遇镜像限制、healthcheck.sh 不存在、60s readiness delay

**第四阶段**：自写 Chart templates，零外部依赖 → `dtk deploy` 一键搞定

---

## ⚠️ 本次踩坑完整记录

### 坑 1：deployment.yaml 缺少 env 渲染
**现象**：`values.yaml` 里写了 `DATABASE_URL`，pod 里完全没有。  
**根因**：Helm 默认生成的 `deployment.yaml` 模板没有 `{{- with .Values.env }}` 块。  
**解决**：给 deployment.yaml 加 env + envFrom 渲染。  
**教训**：Helm 默认模板是最小骨架，env 注入要自己加。

---

### 坑 2：Bitnami etcd healthcheck.sh 不存在
**现象**：`quay.io/coreos/etcd` 镜像跑 Bitnami chart，readiness probe 报 `no such file or directory`。  
**根因**：Bitnami chart 的 readiness probe 调用 `/opt/bitnami/scripts/etcd/healthcheck.sh`，这个脚本只在 Bitnami 自己的镜像里有，官方镜像没有。  
**解决**：弃用 Bitnami chart，自写 etcd Deployment，probe 改用 `GET /health`。  
**教训**：第三方 chart 和非官方镜像强绑定，换镜像必须同步改 probe。

---

### 坑 3：etcd initial-cluster 和 initial-advertise-peer-urls 不匹配
**现象**：etcd CrashLoopBackOff，报 `--initial-cluster has default=http://localhost:2380 but missing from --initial-advertise-peer-urls=http://<pod-ip>:2380`。  
**根因**：etcd 启动时 `--initial-advertise-peer-urls` 默认用 pod IP，但 `--initial-cluster` 默认是 `localhost:2380`，两者不一致导致 cluster 初始化失败。  
**解决**：command 里显式加：
```
--initial-advertise-peer-urls=http://0.0.0.0:2380
--initial-cluster=default=http://0.0.0.0:2380
```
**教训**：etcd 单节点部署必须显式指定这两个参数，不能依赖默认值。

---

### 坑 4：K8s Service 注入的环境变量污染 etcd
**现象**：etcd 日志里满是 `unrecognized environment variable: ETCD_SERVICE_PORT_CLIENT=2379` 之类的警告。  
**根因**：K8s 会把同 namespace 的所有 Service 以 `<SERVICE>_PORT`、`<SERVICE>_SERVICE_HOST` 等形式注入到每个 pod 的环境变量里。etcd 把以 `ETCD_` 开头的环境变量都当成配置，所以 K8s 注入的 `ETCD_SERVICE_PORT_CLIENT` 被 etcd 读到了。  
**解决**：这些是 warn 不是 error，不影响运行，忽略即可。生产环境可以用 `enableServiceLinks: false` 禁止 K8s 注入 Service 环境变量。  
**教训**：K8s Service 环境变量注入是默认行为，命名有冲突风险，敏感服务考虑关闭。

---

### 坑 5：duplicate migration file
**现象**：`migrate source: failed to init driver: duplicate migration file: 000002_password_reset_tokens.down.sql`。  
**根因**：`migrations/` 目录里同时存在 `000002_add_email_to_users` 和 `000002_password_reset_tokens` 两个版本号为 2 的文件。前者是早期遗留，后者是正确的。  
**解决**：删掉 `000002_add_email_to_users.up/down.sql` 和误入的 `user_queries.sql`。  
**教训**：migration 文件的版本号必须全局唯一，不能有重复。新建迁移前先 `ls migrations/` 确认最大版本号。

---

### 坑 6：postgres 被 dtk deploy 意外删除
**现象**：手动 `helm install postgres bitnami/postgresql` 的 release 在某次 `dtk deploy` 后消失。  
**根因**：`dtk deploy` 跑 `helm upgrade --install blitz`，这是另一个 release，不会动 postgres release。但 postgres 消失的原因是 Bitnami 限制，之前的 release 因为 PVC 密码不匹配问题被手动删除过。  
**解决**：把 postgres 纳入 blitz chart 的 templates，不再依赖独立 release。  
**教训**：基础设施组件必须和业务服务在同一个 chart 里管理，独立 release 容易出现生命周期不一致。

---

## 文件变动清单

```
新增：
- deployments/blitz/templates/postgres-statefulset.yaml
- deployments/blitz/templates/etcd-deployment.yaml

修改：
- deployments/blitz/templates/deployment.yaml     ← 加 env/envFrom + initContainers
- deployments/blitz/templates/wallet-service-deployment.yaml ← initContainers 改 service name
- deployments/blitz/Chart.yaml                    ← dependencies: []
- deployments/blitz/values.yaml                   ← DATABASE_URL=postgres:5432, ETCD_ENDPOINTS=etcd:2379
- internal/db/migrations/                              ← 删除重复的 000002_add_email_to_users

删除：
- deployments/blitz/templates/deployment.yaml（nginx 占位）
- deployments/blitz/charts/*.tgz
- internal/db/migrations/000002_add_email_to_users.up.sql
- internal/db/migrations/000002_add_email_to_users.down.sql
- internal/db/migrations/user_queries.sql
```

---

## 当前 K8s 架构

```
blitz namespace
├── postgres (StatefulSet)
│   ├── image: postgres:16-alpine
│   ├── PVC: 1Gi
│   └── probe: pg_isready
├── etcd (Deployment)
│   ├── image: quay.io/coreos/etcd:v3.5.14
│   ├── storage: emptyDir
│   └── probe: GET /health
├── bitcoind (Deployment)
│   └── image: 自定义
├── geth-rpc (Deployment)
│   └── image: 自定义
└── wallet-service (Deployment)
    ├── image: qingchun22/wallet-service-arm64:v0.1.4
    ├── initContainers: wait-postgres + wait-etcd
    ├── golang-migrate: 自动执行迁移
    └── env: DATABASE_URL + ETCD_ENDPOINTS
```

---

## 遗留问题

- ETH geth `403 Forbidden: invalid host specified` — geth 配置问题，不影响主流程
- wallet-service 环境变量明文写在 values.yaml — 生产应改用 K8s Secret
- SMTP 配置未注入 K8s — 忘记密码功能在 K8s 环境下不可用

---

## 历史快照

```
snapshots/
├── SNAPSHOT-blitz-2026-03-23-auth-migrate.md
├── SNAPSHOT-blitz-2026-03-23-k8s-deploy.md
└── SNAPSHOT-blitz-2026-03-24-fully-deployed.md   ← 本次 🏆
```
