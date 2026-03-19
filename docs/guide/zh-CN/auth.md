# 认证授权指南

> web3-blitz 的用户系统、JWT 认证和 refresh token 机制说明。

---

## 整体流程

```
注册 → 登录 → 获得 access_token + refresh_token
                    │
                    ├─ access_token（1小时）→ 调用受保护接口
                    │
                    └─ refresh_token（7天）→ /api/v1/refresh → 换新 token
                                                │
                                                └─ /api/v1/logout → 撤销
```

---

## 用户注册

```bash
POST /api/v1/register
```

```json
// 请求
{"username": "alice", "password": "password123"}

// 响应
{"id": 1, "username": "alice"}
```

密码要求：不少于 8 位，bcrypt DefaultCost 加密存储。

---

## 用户登录

```bash
POST /api/v1/login
```

```json
// 请求
{"username": "alice", "password": "password123"}

// 响应
{
  "access_token":  "eyJhbGci...",
  "refresh_token": "199d90be...",
  "user_id": 1,
  "username": "alice"
}
```

---

## 调用受保护接口

需要 JWT 的接口（当前只有 `/api/v1/withdraw`）：

```bash
curl -X POST http://localhost:2113/api/v1/withdraw \
  -H "Authorization: Bearer eyJhbGci..." \
  -H "Content-Type: application/json" \
  -d '{...}'
```

**注意**：Bearer 后面跟的是 `access_token`，不是 `refresh_token`。

---

## Token 刷新

access_token 过期（1小时）后用 refresh_token 换新的：

```bash
POST /api/v1/refresh
```

```json
// 请求
{"refresh_token": "199d90be..."}

// 响应
{
  "access_token":  "eyJhbGci...",
  "refresh_token": "ff4da9ac...",
  "user_id": 1,
  "username": "alice"
}
```

**旋转策略**：每次刷新都会生成新的 refresh_token，旧的立即撤销。这样即使 refresh_token 泄露，攻击者使用后你会立刻察觉（下次刷新会报 401）。

---

## 退出登录

```bash
POST /api/v1/logout
```

```json
// 请求
{"refresh_token": "ff4da9ac..."}

// 响应
{"message": "已退出登录"}
```

撤销 refresh_token，下次刷新会返回 401。access_token 不在服务端存储，只能等待自然过期（1小时内仍有效）。

---

## JWT 结构

Payload 包含：
```json
{
  "user_id":  1,
  "username": "alice",
  "exp": 1773972471,
  "iat": 1773886071
}
```

签名算法：HS256，密钥从 `JWT_SECRET` 环境变量读取。

---

## 安全设计

**access_token 短期（1小时）**：泄露损失可控，不需要服务端维护黑名单。

**refresh_token 存 DB**：长期凭证，必须支持主动撤销，需要服务端验证。

**旋转策略**：每次刷新换新 token，旧 token 立即失效。refresh_token 被盗用时，攻击者刷新后你的 token 失效，再刷新时报错，从而发现异常。

**生产环境 JWT_SECRET**：
```bash
# 生成强密钥
openssl rand -hex 32

# K8s Secret 注入
env:
  - name: JWT_SECRET
    valueFrom:
      secretKeyRef:
        name: wallet-secret
        key: jwt-secret
```

---

## 当前受保护接口

| 接口 | 说明 |
|------|------|
| POST /api/v1/withdraw | 发起提币 |

后续计划保护：deposits、withdrawals、balance/total 查询接口。
