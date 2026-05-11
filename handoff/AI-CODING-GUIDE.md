# AI 编码指南

> 本文档写给协助开发的 AI。
> 项目骨架由 `kp init` 生成，请严格在框架内填充业务逻辑，不要重构脚手架结构。

---

## 一、项目框架总览

```
blitz/
├── cmd/blitz/
│   └── main.go              # 服务入口，只做初始化和启动，不写业务逻辑
├── internal/
│   ├── api/
│   │   ├── handler.go       # HTTP handler，业务入口
│   │   └── server.go        # 路由注册，HTTP 服务启动
│   ├── auth/
│   │   ├── auth.go          # JWT 签发/验证
│   │   ├── middleware.go    # 认证中间件
│   │   └── rbac.go          # 权限控制
│   ├── db/
│   │   ├── connect.go       # 数据库连接初始化
│   │   └── migrations/      # embed.FS 加载的 SQL 文件目录
│   ├── metrics/
│   │   └── metrics.go       # Prometheus 指标定义
│   └── pkg/
│       └── code/            # 业务错误码
├── configs/
│   ├── project.env          # 部署配置（VERSION、KUBE_* 等）
│   ├── components.yaml      # kp AI 规划依据
│   ├── resources.yaml       # controller 监控的 K8s 资源
│   └── system.yaml          # KubePivot 控制器配置（etcd/调度/缓存参数）
├── deployments/blitz/  # Helm chart，含 postgres、etcd、业务服务
├── build/docker/blitz/# Dockerfile + build.sh
├── migrations/              # SQL 迁移文件（golang-migrate 格式）
└── handoff/
    ├── HANDOFF.md           # 项目上下文，写给下一个 AI
    └── AI-CODING-GUIDE.md   # 本文件
```

---

## 二、业务逻辑该填在哪里

### API Handler

`internal/api/handler.go` 是业务入口，新增接口在这里加 handler 函数：

```go
// 示例：新增一个查询接口
func (h *Handler) GetUser(w http.ResponseWriter, r *http.Request) {
    // 1. 解析请求参数
    // 2. 调用业务逻辑（建议抽到 internal/service/ 或直接写在这里）
    // 3. 统一用 h.respond(w, code, data) 返回
}
```

路由注册在 `internal/api/server.go`，新增接口后在这里加一行。

### 数据库操作

`internal/db/` 负责连接管理，具体查询建议按业务模块拆包，例如：

```
internal/
└── db/
    ├── connect.go           # 已有，不动
    ├── migrations/          # 已有，不动
    ├── user.go              # 新增：用户相关查询
    └── wallet.go            # 新增：钱包相关查询
```

### 数据库迁移

`migrations/` 目录存放 SQL 文件，**严格遵守 golang-migrate 命名格式**：

```
migrations/
├── 000001_init_schema.up.sql
├── 000001_init_schema.down.sql
├── 000002_add_users.up.sql
└── 000002_add_users.down.sql
```

迁移在服务启动时自动执行，不需要手动触发。

### 扩展业务模块

复杂业务建议在 `internal/` 下按模块新增包，例如：

```
internal/
├── service/                 # 业务逻辑层（可选）
│   ├── user.go
│   └── wallet.go
└── model/                   # 数据模型（可选）
    └── user.go
```

包名和目录名保持一致，不要在 `internal/` 外新增业务代码。

---

## 三、不要动的地方

以下文件和目录由 kp 框架管理，**不要修改**，否则会破坏部署流程：

| 文件/目录 | 原因 |
|---|---|
| `configs/project.env` | kp deploy 读取，手动改会导致部署参数错乱 |
| `configs/components.yaml` | AI 规划依据，改了会影响资源估算 |
| `configs/resources.yaml` | controller 监控配置，改了需要重新 deploy |
| `configs/system.yaml` | KubePivot 控制器参数（可通过 env 覆盖，文件不要乱改）|
| `deployments/` 目录结构 | Helm chart 骨架，新增配置只改 `values.yaml` |
| `scripts/make-rules/deploy.mk` | kp 部署流程，不要改 helm upgrade 参数 |
| `internal/db/migrations/` | embed.FS 挂载点，目录不能改名 |
| `cmd/blitz/main.go` | 只做初始化，业务逻辑不要写在这里 |

---

## 四、加新功能的标准姿势

### 加一个新 API 接口

```
1. internal/api/handler.go   → 加 handler 函数
2. internal/api/server.go    → 注册路由
3. internal/db/{module}.go   → 加数据库操作（如需要）
4. migrations/               → 加迁移文件（如需要改表结构）
```

### 加一张新数据库表

```
1. 新建 migrations/00000N_add_{table}.up.sql   → 建表 SQL
2. 新建 migrations/00000N_add_{table}.down.sql → 回滚 SQL
3. internal/db/{module}.go                     → 加查询函数
```

序号 N 必须连续递增，不能跳号。

### 加环境变量

```
1. deployments/blitz/values.yaml  → 在 env 里加
2. configs/project.env               → 本地开发时加（不提交敏感值）
```

不要硬编码配置值，所有配置通过环境变量注入。

### 加 Prometheus 指标

```
1. internal/metrics/metrics.go  → 定义新指标
2. internal/api/handler.go      → 在对应 handler 里埋点
```

---

## 五、代码风格约束

**日志**：统一用 `log/slog`，不用 `log` 包。结构化输出，key 用小写英文：

```go
// ✓ 正确
slog.Info("用户登录", "user_id", uid, "ip", ip)
slog.Error("数据库查询失败", "err", err, "table", "users")

// ✗ 不要用
log.Printf("user %d login", uid)
fmt.Println("error:", err)
```

**错误处理**：错误向上传递，加上下文信息，不要在中间层吞掉：

```go
// ✓ 正确
if err != nil {
    return fmt.Errorf("查询用户失败: %w", err)
}

// ✗ 不要这样
if err != nil {
    log.Println(err) // 吞掉了
    return nil
}
```

**包组织**：严格分包，不要跨层调用。依赖方向：

```
api → service → db
api → auth
api → metrics
```

`db` 包不能 import `api` 包，`auth` 包不能 import `db` 包，以此类推。

**模块路径**：项目模块路径是 `github.com/Ixecd/blitz`，import 内部包时用完整路径：

```go
import (
    "github.com/Ixecd/blitz/internal/db"
    "github.com/Ixecd/blitz/internal/auth"
)
```

---

## 六、改了代码要同步改哪里

| 改动类型 | 需要同步的地方 |
|---|---|
| 新增环境变量 | `deployments/blitz/values.yaml` 的 `env` 字段 |
| 改了服务端口 | `values.yaml` 的 `service.port` + liveness/readiness probe 端口 |
| 加了新的 K8s 资源需要监控 | `configs/resources.yaml` 加一行，然后 `kp deploy` |
| 改了数据库表结构 | 新增迁移文件，不要修改已有迁移文件 |
| 要发布新版本 | 运行 `kp release --version vX.Y.Z`，不要手动改 VERSION |
| 同步框架文件 | `kp sync`，KubePivot 升级后同步 Makefile/scripts/configs |
| 性能基准 | `kp bench all`，跑全量 KVCache/调度/内存基准 |

---

> 遇到不确定的地方，先看 `handoff/HANDOFF.md` 了解项目背景，再动手。
> 不确定该不该改某个文件，默认答案是：不改，先问。
