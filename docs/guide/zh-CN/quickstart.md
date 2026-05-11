# 快速开始

> 5 分钟内在本地跑起 blitz。

## 前置条件

| 工具 | 版本要求 | 安装 |
|------|----------|------|
| Go | >= 1.24 | [golang.org](https://golang.org) |
| Docker | latest | [docker.com](https://docker.com) |
| bitcoin-cli | >= 28.0 | `brew install bitcoin` |
| dtk | latest | `go install github.com/Ixecd/blitz/cmd/dtk@latest` |

---

## 第一步：克隆项目
```bash
git clone https://github.com/Ixecd/blitz.git
cd blitz
go mod tidy
```

---

## 第二步：启动依赖
```bash
docker compose up -d
```

启动 bitcoind（regtest）和 geth（dev mode）。

验证：
```bash
# bitcoind 是否正常
bitcoin-cli -regtest getblockchaininfo

# geth 是否正常
curl -s -X POST http://localhost:8545 \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}'
```

---

## 第三步：初始化 BTC 钱包
```bash
bitcoin-cli -regtest createwallet "blitz_wallet"
bitcoin-cli -regtest -rpcwallet=blitz_wallet -generate 101
bitcoin-cli -regtest -rpcwallet=blitz_wallet getbalance
# 输出：50.00000000
```

---

## 第四步：启动服务
```bash
go run cmd/wallet-service/main.go
```

期望输出：
```
✅ 数据库已连接
✅ 从DB恢复了 0 个充值地址
🚀 Wallet Core 服务已启动
🔍 Deposit Watcher 已启动
📦 从块高 102 开始监听
📡 API 服务已启动: http://localhost:2113
```

---

## 第五步：测试充值全流程
```bash
# 1. 生成充值地址
curl -X POST http://localhost:2113/api/v1/address \
  -H "Content-Type: application/json" \
  -d '{"user_id":"test001","chain":"btc"}'

# 输出：
# {"address":"bcrt1q...","path":"m/44'/0'/0'/0/...","user_id":"test001"}

# 2. 往充值地址打币
bitcoin-cli -regtest -rpcwallet=blitz_wallet \
  sendtoaddress "bcrt1q..." 0.1

# 3. 挖块确认
bitcoin-cli -regtest -rpcwallet=blitz_wallet -generate 1

# 4. 观察服务日志
# 💰 检测到充值！userID=test001 amount=0.100000
# ✅ deposit已写入DB
```

---

## 第六步（可选）：部署到 K8s

配置 `configs/project.env`：
```ini
REGISTRY_PREFIX=<你的DockerHub用户名>
ARCH=arm64   # M芯片用arm64，Intel用amd64
VERSION=v0.1.1
```

一键部署：
```bash
dtk deploy
```

验证：
```bash
kubectl get pods -n blitz
# NAME                             READY   STATUS    
# bitcoind-xxx                     1/1     Running   
# geth-rpc-xxx                     1/1     Running   
# wallet-service-xxx               1/1     Running   
```

---

## 下一步

- [完整部署指南](deploy.md)
- [dtk 使用指南](dtk.md)
- [wallet-core 设计文档](../../design/wallet-core.md)