# GitHub Actions CI 流程文档

> web3-blitz 的持续集成配置说明。

---

## 触发条件

```yaml
on:
  push:
    branches: [ main ]
  pull_request:
    branches: [ main ]
```

- **push main**：直接推送到 main 分支时触发
- **PR to main**：提交 PR 时触发，CI 通过才能合并

---

## 流程

```
push / PR
    │
    ▼
Set up Go 1.25
    │
    ▼
Init DB schema（postgres:16-alpine）
    │
    ▼
go build ./...
    │
    ▼
go vet ./...
    │
    ▼
go test ./...
    │
    ▼
docker build
```

---

## 测试分层

| 类型 | 目录 | 触发方式 | 说明 |
|------|------|---------|------|
| integration | `test/integration/` | 普通 `go test` | 用 httptest，无需外部服务，CI 自动跑 |
| e2e | `test/e2e/` | `-tags e2e` | 需服务启动，CI 不跑 |
| smoke | `test/smoke/` | `-tags e2e` | 需服务启动，CI 不跑 |

e2e 和 smoke 测试文件顶部有 `//go:build e2e` tag，普通 `go test ./...` 自动跳过。

本地跑 e2e：
```bash
# 先启动服务
go run cmd/wallet-service/main.go

# 再跑 e2e
go test -tags e2e ./test/e2e/... -v -timeout 120s
```

---

## CI 环境

CI 里自动启动 `postgres:16-alpine` 服务：

```yaml
services:
  postgres:
    image: postgres:16-alpine
    env:
      POSTGRES_USER: blitz
      POSTGRES_PASSWORD: blitz
      POSTGRES_DB: blitz
    options: >-
      --health-cmd pg_isready
      --health-interval 5s
```

schema 初始化：
```yaml
- name: Init DB schema
  run: PGPASSWORD=blitz psql -h localhost -U blitz -d blitz -f internal/db/schema.sql
```

---

## 常见问题

**go vet 报 non-constant format string**

`g.Printf(buf.String())` 改成 `g.Printf("%s", buf.String())`。

**docker build 报 go.mod requires go >= 1.25.0**

Dockerfile 基础镜像和 go.mod 版本不一致，统一用 `golang:1.25-alpine`。

**e2e 测试在 CI 里失败**

e2e/smoke 测试需要服务运行，加 `//go:build e2e` tag 让 CI 跳过。

**分支保护设置**

GitHub → Settings → Branches → Add rule：
- 勾选 `Require status checks to pass before merging`
- 选择 `Build & Test` job
- 防止 CI 失败的代码合并进 main
