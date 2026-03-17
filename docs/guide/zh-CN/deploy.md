# 部署指南

## 前置条件

- Docker Desktop / OrbStack
- kubectl
- helm
- dtk（`go install github.com/Ixecd/dev-toolkit/cmd/dtk@latest`）

## 本地开发

### 1. 启动依赖
```bash
docker compose up -d
```

启动 bitcoind（regtest）和 geth（dev mode）。

### 2. 初始化 BTC 钱包
```bash
bitcoin-cli -regtest createwallet "blitz_wallet"
bitcoin-cli -regtest -rpcwallet=blitz_wallet -generate 101
```

### 3. 启动服务
```bash
go run cmd/wallet-service/main.go
```

### 4. 验证
```bash
# 生成充值地址
curl -X POST http://localhost:2113/api/v1/address \
  -H "Content-Type: application/json" \
  -d '{"user_id":"test001","chain":"btc"}'

# 往充值地址打币
bitcoin-cli -regtest -rpcwallet=blitz_wallet sendtoaddress "bcrt1q..." 0.1

# 挖块确认
bitcoin-cli -regtest -rpcwallet=blitz_wallet -generate 1

# 观察服务日志，应该看到：
# 💰 检测到充值！
# ✅ deposit已写入DB
```

---

## K8s 部署

### 1. 配置

编辑 `configs/project.env`：
```ini
PROJECT_NAME=web3-blitz
KUBE_NAMESPACE=web3-blitz
REGISTRY_PREFIX=<你的DockerHub用户名>
ARCH=arm64        # 或 amd64
VERSION=v0.1.1
```

编辑 `configs/components.yaml`，确认要部署的服务：
```yaml
components:
  - name: wallet-service
    port: 2113
    image: wallet-service
```

### 2. 一键部署
```bash
dtk deploy
```

流程：
1. AI 规划资源（replicas/cpu/memory）
2. docker build + push 镜像
3. helm upgrade --install
4. kubectl set image + rollout

### 3. 验证
```bash
# 查看 pod 状态
kubectl get pods -n web3-blitz

# 查看 wallet-service 日志
kubectl logs -n web3-blitz deployment/wallet-service

# 期望输出：
# ✅ 数据库已连接
# 🔍 Deposit Watcher 已启动
# 📦 从块高 0 开始监听
```

### 4. 访问服务
```bash
kubectl port-forward -n web3-blitz svc/wallet-service 2113:2113
curl http://localhost:2113/metrics
```

---

## 常见问题

**bitcoind 余额为 0**
regtest 下 coinbase 需要 101 个确认才能花费，先挖够块：
```bash
bitcoin-cli -regtest -rpcwallet=blitz_wallet -generate 101
```

**geth CrashLoopBackOff**
K8s 自动注入 `GETH_PORT` 环境变量导致 geth 解析失败，确保 geth deployment 里有：
```yaml
spec:
  template:
    spec:
      enableServiceLinks: false
```

**镜像拉不到**
K8s 缓存了旧镜像，改 `project.env` 里的 `VERSION` 后重新 `dtk deploy`。

**网络连接不上 Docker Hub**
build 时加参数：
```bash
dtk deploy  # 会自动用 DOCKER_PULL=false
```
或者提前手动拉好基础镜像：
```bash
docker pull golang:1.24-alpine
docker pull alpine:3.20
```