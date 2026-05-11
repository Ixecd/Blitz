# 项目交接文档

> 写给下一个 Claude
> 日期：2026-05-11
> 作者：<!-- 填写作者 -->

---

## 写在前面

<!-- 描述你的工作风格和偏好，Claude 会据此调整协作方式。例如：
- 设计优先，代码其次。不要上来就写代码，先对齐设计再动手
- 每完成一个里程碑：commit → tag → SNAPSHOT → 更新 TODO
- 喜欢被推 back，不喜欢被一味认同
- 代码风格偏好（日志库、包组织方式等）
-->

---

## 一、项目概览

**仓库**：github.com/Ixecd/blitz
**当前版本**：<!-- vX.Y.Z，与 configs/project.env 中 VERSION 保持一致 -->
**定位**：<!-- 一句话说清楚这个项目是什么、解决什么问题 -->

**命令全览**：

```
kp init       --name blitz --module <module> [--with-frontend]
kp deploy     [--namespace] [--context] [--kubeconfig] [--dry-run]
kp resume     # 从中断点恢复
kp rollback   # 手动触发 helm rollback
kp release    --version vX.Y.Z [--deploy]
```

**目录结构**：

```
blitz/
├── cmd/
│   └── blitz/
│       └── main.go          # 服务入口
├── internal/                # 业务逻辑（按模块拆分）
├── configs/
│   ├── project.env          # 项目配置（VERSION、KUBE_*、ETCD_ENDPOINTS 等）
│   ├── components.yaml      # kp 部署组件声明
│   └── resources.yaml       # controller 监控的 K8s 资源列表
├── deployments/
│   └── blitz/        # 自包含 Helm chart
│       ├── Chart.yaml
│       ├── values.yaml
│       └── templates/
├── migrations/              # SQL 迁移文件（golang-migrate）
├── snapshots/               # 里程碑快照归档
├── handoff/
│   └── HANDOFF.md           # 本文件
└── docs/                    # 设计文档
```

---

## 二、架构设计

<!-- 描述核心架构决策，重点说「为什么这么设计」而不是「是什么」。
Claude 理解了设计意图，才不会给出破坏架构的建议。
-->

---

## 三、当前状态

<!-- 描述项目现在处于什么阶段，哪些功能已经可用，哪些在开发中。
让 Claude 对「完成度」有准确认知，避免重复已完成的工作。
-->

---

## 四、已知问题

<!-- 列出已知但暂未修复的 Bug 或设计缺陷，注明优先级。
格式建议：
1. [P0] 问题描述 — 根因（如果已知）
2. [P1] 问题描述
-->

---

## 五、接下来要做的事

<!-- 按优先级列出下一步计划。
写得足够具体，避免「优化性能」这种模糊描述，
要写「把 handler.go 按 auth/wallet/admin 拆分」。
-->

### P0（最优先）

### P1

---

## 六、常用命令速查

```bash
# 部署
kp deploy

# 查看 pods
kubectl get pods -n blitz

# 查看服务日志
kubectl logs -n blitz deployment/blitz

# port-forward 本地调试
kubectl port-forward -n blitz deployment/blitz <local-port>:<container-port>

# 查看 helm 历史
helm history blitz -n blitz

# 运行测试
go test ./...
```

---

## 七、快照归档位置

`snapshots/` 目录下按日期和里程碑命名，格式：

```
SNAPSHOT-{项目}-{日期}-{里程碑}.md
```

---

## 八、致下一个 Claude

<!-- 写给 Claude 的话。可以说说这个项目对你的意义，
或者特别需要 Claude 注意的地方。这部分会影响 Claude 的协作态度。
-->
