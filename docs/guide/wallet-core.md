# Wallet Core 设计文档

> 当前版本：v0.1.1 | 状态：开发中

## 概述

wallet-service 是 web3-blitz 的核心服务，负责：

- HD 钱包地址派生
- 充值监控（Deposit Watcher）
- 链上数据持久化
- REST API 对外暴露

## 架构
```
HTTP Client
     │
     ▼
┌─────────────────────────────┐
│        wallet-service       │
│                             │
│  /api/v1/address  ──────────────► HDWallet (BIP44)
│  /api/v1/balance  ──────────────► BTC/ETH RPC
│  /metrics         ──────────────► Prometheus
│                             │
│  AddressRegistry(Mem)       │
│       ▲         │           │
│       │         ▼           │
│  DepositWatcher  ──────────────► bitcoind (扫块)
│                             │
│       │                     │
│       ▼                     │
│     SQLite (blitz.db)       │
└─────────────────────────────┘
```

## 核心组件

### HDWallet

基于 BIP44 标准派生地址。

- BTC 路径：`m/44'/0'/0'/0/<index>`
- ETH 路径：`m/44'/60'/0'/0/<index>`
- index 由 userID hash 生成，确保同一用户每次得到相同地址
- 底层库：`go-bip32`
```go
// 根据 userID 派生充值地址
resp, err := btcWallet.GenerateDepositAddress(ctx, userID, types.ChainBTC)
```

### AddressRegistry

内存中维护 `address → userID` 的映射表，服务启动时从 DB 恢复。
```
服务启动 → 从 deposit_addresses 表加载所有地址 → 填充 registry
生成新地址 → 写 DB → 同时注册到 registry
```

**线程安全**：读写锁保护。

### DepositWatcher

每 3 秒扫一次新块，检测充值到账。
```
每3s → GetBlockChainInfo → 发现新块
     → GetBlockVerboseTx → 遍历所有交易输出
     → 匹配 AddressRegistry → 命中则推入 deposits channel
     → 消费 goroutine → 写入 deposits 表
```

**当前支持**：BTC（regtest）
**待支持**：ETH

### 数据库

SQLite + sqlc，三张表：

| 表名 | 说明 |
|------|------|
| deposit_addresses | 已生成的充值地址 |
| deposits | 已确认的充值记录 |
| withdrawals | 提币记录（待实现） |

## API

### POST /api/v1/address

生成充值地址。

**请求：**
```json
{
  "user_id": "user001",
  "chain": "btc"
}
```

**响应：**
```json
{
  "address": "bcrt1q...",
  "path": "m/44'/0'/0'/0/2872464479",
  "user_id": "user001"
}
```

### GET /api/v1/balance

查询地址余额。
```
GET /api/v1/balance?address=bcrt1q...&chain=btc
```

**响应：**
```json
{
  "address": "bcrt1q...",
  "balance": 0.1,
  "chain": "btc"
}
```

### GET /metrics

Prometheus 指标端点。

## 待实现

- [ ] ETH Deposit Watcher
- [ ] 提币接口（/api/v1/withdraw）
- [ ] 用户余额查询（/api/v1/deposits）
- [ ] 多签支持
- [ ] PostgreSQL 迁移