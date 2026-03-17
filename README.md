# web3-blitz ⚡

> 基于 Go + K8s 的 Web3 交易所基础设施，支持 BTC/ETH 充值监控、HD 钱包派生、链上数据持久化。

## 🏗️ 架构
```
┌─────────────────────────────────────────────┐
│                  K8s Cluster                │
│                                             │
│  ┌──────────────┐     ┌──────────────────┐  │
│  │   bitcoind   │     │     geth-rpc     │  │
│  │  (regtest)   │     │   (--dev mode)   │  │
│  └──────┬───────┘     └────────┬─────────┘  │
│         │                      │            │
│  ┌──────▼──────────────────────▼─────────┐  │
│  │           wallet-service              │  │
│  │    HD派地址 │ Deposit Watcher │ REST   │  │
│  │           SQLite (blitz.db)           │  │
│  └───────────────────────────────────────┘  │
└─────────────────────────────────────────────┘
```

## ✨ Features

- **HD钱包**：BIP44路径派生，支持BTC/ETH充值地址生成
- **Deposit Watcher**：实时扫块监听充值，自动入账
- **持久化**：SQLite + sqlc，服务重启不丢状态
- **可观测**：Prometheus metrics + Grafana dashboard
- **云原生**：一键`dtk deploy`部署到K8s

## 🚀 快速开始

### 本地开发
```bash
# 启动所有依赖（bitcoind + geth）
docker compose up -d

# 启动服务
go run cmd/wallet-service/main.go
```

### 停止依赖
```bash
docker compose down
```

### 一键部署到K8s
```bash
# 确保configs/project.env配置正确
dtk deploy
```

### 生成充值地址
```bash
curl -X POST http://localhost:2113/api/v1/address \
  -H "Content-Type: application/json" \
  -d '{"user_id":"user001","chain":"btc"}'
```

### 查询余额
```bash
curl "http://localhost:2113/api/v1/balance?address=bcrt1q...&chain=btc"
```

## 📦 服务说明

| 服务 | 端口 | 说明 |
|------|------|------|
| wallet-service | 2113 | 核心钱包服务，REST API |
| bitcoind | 18443 | BTC节点 (regtest) |
| geth-rpc | 8545 | ETH节点 (dev mode) |
| Prometheus | 2112 | 指标采集 |
| Grafana | 3000 | 可视化面板 |

## 🗂️ 项目结构
```
web3-blitz/
├── cmd/
│   ├── wallet-service/    # 核心服务入口
│   └── mining/            # 挖矿工具
├── internal/
│   ├── wallet/            # HD钱包、Watcher、BTC/ETH实现
│   └── db/                # sqlc生成的DB层
├── deployments/           # Helm Chart
├── build/docker/          # Dockerfile
├── configs/               # 配置文件
└── docs/                  # 文档
```

## ⚙️ 配置

`configs/project.env`：
```ini
PROJECT_NAME=web3-blitz
KUBE_NAMESPACE=web3-blitz
REGISTRY_PREFIX=qingchun22
ARCH=arm64
VERSION=v0.1.1
```

`configs/components.yaml`：
```yaml
components:
  - name: wallet-service
    port: 2113
    image: wallet-service
```

## 📚 文档

- [wallet-core 设计文档](docs/wallet-core.md)
- [部署指南](docs/guide/deploy.md)
- [dtk 使用指南](docs/guide/dtk.md)

## 📄 License

MIT