# dtk 使用指南

> dtk（blitz）是 blitz 的脚手架工具，提供项目初始化和一键部署能力。

## 安装
```bash
go install github.com/Ixecd/blitz/cmd/dtk@latest
```

验证安装：
```bash
dtk --help
```

---

## dtk init

初始化一个新项目，生成完整的项目骨架。

### 用法
```bash
dtk init --name <项目名> --module <go模块路径> [flags]
```

### 参数

| 参数 | 说明 | 示例 |
|------|------|------|
| `--name` | 项目名（小写，必填） | `my-service` |
| `--module` | Go 模块路径（必填） | `github.com/you/my-service` |
| `--output` | 输出目录（默认 `./<name>`） | `~/projects/my-service` |
| `--template` | 自定义模板目录 | `/path/to/templates` |
| `--force` | 允许覆盖非空目录 | - |

### 示例
```bash
dtk init --name wallet-svc --module github.com/me/wallet-svc
cd wallet-svc
go mod tidy
```

### 生成的文件结构
```
wallet-svc/
├── cmd/wallet-svc/main.go        # HTTP 服务入口（:8080 /healthz）
├── configs/
│   ├── components.yaml           # AI 部署规划输入
│   └── project.env               # 项目环境变量
├── deployments/wallet-svc/       # Helm Chart
│   ├── Chart.yaml
│   ├── values.yaml
│   └── templates/
├── build/docker/wallet-svc/      # Dockerfile（多阶段构建）
│   ├── Dockerfile
│   └── build.sh
├── scripts/make-rules/           # Makefile 规则集
├── Makefile
└── .gitignore
```

---

## dtk deploy

一键构建镜像、推送到 Registry、部署到 K8s。

### 用法
```bash
dtk deploy [flags]
```

### 参数

| 参数 | 说明 | 默认值 |
|------|------|--------|
| `--components` | 组件配置文件路径 | `configs/components.yaml` |
| `--namespace` | K8s 命名空间 | 读取 `project.env` |
| `--context` | K8s context | 读取 `project.env` |
| `--dry-run` | 只打印规划，不实际部署 | `false` |

### 部署流程
```
dtk deploy
    │
    ├─ 1. 读取 configs/components.yaml
    ├─ 2. AI 规划资源（replicas/cpu/memory/storage）
    ├─ 3. make deploy.build（docker build 每个服务）
    ├─ 4. make deploy.push（docker push 到 Registry）
    ├─ 5. helm upgrade --install（安装或更新 Chart）
    ├─ 6. kubectl set image（更新镜像）
    ├─ 7. kubectl rollout status（等待滚动更新完成）
    └─ 8. kubectl set resources（设置 CPU/Memory 限制）
```

### 示例
```bash
# 查看规划，不实际部署
dtk deploy --dry-run

# 部署到默认命名空间
dtk deploy

# 部署到指定命名空间
dtk deploy --namespace production
```

---

## 配置文件

### configs/project.env

控制部署行为的核心配置：
```ini
PROJECT_NAME=blitz          # 项目名，对应 Helm release 名
KUBE_NAMESPACE=blitz        # K8s 命名空间
KUBE_CONTEXT=                    # K8s context（留空使用当前）
REGISTRY_PREFIX=qingchun22       # Docker Registry 前缀
ARCH=arm64                       # 目标架构（arm64/amd64）
VERSION=v0.1.1                   # 镜像版本号
```

> 每次发布新版本，修改 `VERSION` 后执行 `dtk deploy` 即可。

### configs/components.yaml

定义需要部署的服务：
```yaml
components:
  - name: wallet-service    # 服务名（对应 build/docker/<name>/）
    port: 2113              # 服务端口
    image: wallet-service   # 镜像名（留空则跳过 build/push）
```

> `image` 字段留空或注释掉该服务，deploy 时会跳过该服务的构建和部署。

---

## Makefile 集成

dtk deploy 底层调用 Makefile，也可以单独使用：
```bash
# 只构建镜像
make deploy.build IMAGES="wallet-service" PLATFORM=linux_arm64

# 只推送镜像
make deploy.push IMAGES="wallet-service" REGISTRY_PREFIX=qingchun22

# 只部署到 K8s（不构建）
make deploy.install
make deploy.run.all
```

---

## 常见问题

**`dtk: command not found`**

确认 Go bin 目录在 PATH 里：
```bash
export PATH=$PATH:$(go env GOPATH)/bin
```

**`找不到项目根目录`**

dtk 通过查找 `Makefile` 来定位项目根目录，确保在项目目录内执行。

**每次 deploy 都重新构建很慢**

镜像 build 利用 Docker 缓存，只有代码变动的层才会重新构建。如需强制重建：
```bash
# 暂不支持通过 dtk 传递，直接用 make
make deploy.build EXTRA_ARGS="--no-cache"
```

**部署后镜像没有更新**

修改 `configs/project.env` 里的 `VERSION` 字段，K8s 通过镜像 tag 判断是否需要更新。