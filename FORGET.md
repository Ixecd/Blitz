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

## 四、低危（已修 ✅）

### 4.1 Refresh token rotate 顺序修复 ✅
- auth.go Refresh: 先创建新 token 再撤销旧 token，防 DB 写入失败导致用户被锁

### 4.2 审计日志 ✅
- 新增 audit 包：JSON Lines 格式审计日志
- 提币 submitted/completed/failed 三事件全程留痕

### 4.3 ETH gas price 上限 ✅
- eth/withdraw.go: maxGasPrice = 200 gwei，超限拒绝并提示稍后重试

---

## 五、HD 种子硬加密（新增 ✅）

### 明文种子退役 ✅
- WALLET_HD_SEED 环境变量不再使用
- 改为 AES-256-GCM 加密文件 + argon2id 密钥派生
- 密码通过交互输入或 SEED_PASSPHRASE 环境变量提供

### 加密工具 ✅
- 新增 cmd/encrypt-seed: 种子文件加密生成
- encrypt-seed <种子hex> <输出路径> → 密码提示 → 0600 权限输出

---

## 六、Sealed Secrets 部署加密（新增 ✅）

### kp secret seal 已接入 ✅
- 6 个密钥（JWT_SECRET / DATABASE_URL / SMTP_USER / SMTP_PASS / ETH_HOT_WALLET_KEY）全部 kubeseal 加密
- 输出 configs/secrets/blitz-sealed.yaml，安全提交 Git
- 部署到集群后 sealed-secrets controller 自动解密为 K8s Secret → Pod env

### 密钥安全双层 ✅
```
开发机  encrypt-seed → AES-256-GCM 文件
CI/Git  kp secret seal → SealedSecret
集群    kubeseal 解封 → K8s Secret → Pod 挂载
```

---

*全部修完，FORGET 清空。*
