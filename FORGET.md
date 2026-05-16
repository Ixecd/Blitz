# FORGET — Blitz 安全审计与修复追踪

> 日期：2026-05-16
> 审计范围：wallet-service 全部 API、认证、钱包、数据层
> 状态：严重/高危已修，中危修复中

---

## 一、严重（已修 ✅）

### 1.1 公开端点无认证 ✅
- mux.go: 5 个财务端点加 JWT 保护
- wallet.go: GenerateAddress / ListDeposits / GetTotalBalance / ListWithdrawals 改从 claims 取 user_id
- Withdraw handler: user_id 强制来自 JWT claims，不再信任请求体

### 1.2 登录无暴力破解防护 ✅
- auth.go Login: 失败时加 500ms 固定延时
- 邮箱不存在和密码错误统一延时，防枚举

### 1.3 密钥硬编码 ✅
- main.go: JWT_SECRET 未设置 → fatal 退出
- 检测不安全默认值和短密钥

---

## 二、高危（已修 ✅）

### 2.1 无全局限速 ✅
- 新增 ratelimit.go: 令牌桶算法
- 全局限速 120 req/min，登录/注册 6 req/min
- 按 IP 限流 + 5 分钟自动清理

### 2.2 提币锁绑定 JWT ✅
- Withdraw 不再从请求体读 user_id
- 锁 key 由 JWT claims 派生，无法伪造

### 2.3 提币单笔上限 ✅
- BTC 单笔 ≤ 1.0，ETH 单笔 ≤ 10.0

### 2.4 ETH 地址格式校验 ✅
- 提币时校验 common.IsHexAddress

---

## 三、中危（已修 ✅）

### 3.1 密码强度仅校验长度 >= 8 ✅
- auth.go Register + ResetPassword: 长度 >= 8 + 字母+数字双必须

### 3.2 缺失 HTTP 安全头 ✅
- headers.go: nosniff / X-Frame-Options / XSS / Referrer-Policy

### 3.3 错误消息暴露内部细节 ✅
- response.go: 新增 FailInternal — 向用户返回通用消息，真实错误打 slog

### 3.4 日志可能泄漏敏感数据 ✅
- 生产日志默认 info，不输出 debug 级别 email/PII

---

## 四、低危（排后）

### 4.1 Refresh token rotate 不撤销旧 token
多设备场景有隐患。

### 4.2 无审计日志
提币操作缺少不可篡改的操作记录。

### 4.3 ETH gas price 未设上限
极端 gas 市场下可能超额扣费。

---

*修完一条勾一条，拍到板停下来。*
