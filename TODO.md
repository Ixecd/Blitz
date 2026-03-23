# TODO — web3-blitz 路线图

> 充提币系统，从开发环境走向可交付。
> 按优先级排列，持续更新。

---

## 🔴 高优先级

### K8s 部署（下次第一件事）
- [ ] `values.yaml` `walletService.env` 加 `DATABASE_URL`，重新 `dtk deploy`
- [ ] 验证 wallet-service pod 正常启动，迁移自动执行
- [ ] 查 `web3-blitz` pod 重启原因

### 工程
- [ ] etcd 旧 registry 数据清理
- [ ] `handler.go` 拆分：auth、wallet、admin 各自独立文件

### 功能
- [ ] 忘记密码限流（同一邮箱 1 分钟内只能发一次）
- [ ] 注册邮箱验证码验证
- [ ] refresh token 自动续期（前端拦截 401 自动刷新）

---

## 🟡 中优先级

### 页面
- [ ] Dashboard 页面按新设计语言重做
- [ ] Deposit（充值）页面重做
- [ ] Withdraw（提币）页面重做
- [ ] Admin 用户列表显示 email 而非 username

### 功能
- [ ] 扫码登录
- [ ] 登出时踢掉所有设备选项

### K8s
- [ ] wallet-service 环境变量改用 K8s Secret（DATABASE_URL、JWT_SECRET、SMTP_PASS 不应明文写在 values.yaml）
- [ ] SMTP 配置注入（SMTP_HOST / SMTP_USER / SMTP_PASS / SMTP_FROM）
- [ ] etcd 部署到 K8s

---

## 🟢 低优先级

- [ ] 提币审核流程（pending → 人工审核 → 广播）
- [ ] 充值地址二维码生成
- [ ] 交易记录导出 CSV
- [ ] 管理后台操作日志
- [ ] 多币种扩展（USDT、TRX 等）

---

## ✅ 已完成

- [x] HD 钱包生成（BTC + ETH）
- [x] 充值地址生成 + 注册表恢复
- [x] BTC / ETH 充值扫块监听
- [x] 确认数检查（ConfirmChecker，etcd 选主）
- [x] 提币接口（分布式锁防重、余额校验、日限额）
- [x] 死信队列（deposit 写入失败兜底）
- [x] JWT 认证 + refresh token 轮转
- [x] RBAC 权限系统（roles / permissions / user_roles）
- [x] 管理后台基础（用户列表、等级升级、限额配置）
- [x] slog 结构化日志（JSON，LOG_LEVEL=debug）
- [x] register / login 改用 email
- [x] 忘记密码全链路（QQ 邮箱 SMTP）
- [x] 注册新账号（Login 页内联切换）
- [x] 记住我（localStorage / sessionStorage 切换）
- [x] golang-migrate + embed.FS（迁移文件打进二进制，启动自动执行）
- [x] server.go → mux.go 改名
- [x] K8s postgres（Bitnami）部署完成

---

> 每完成一项，移到 ✅ 已完成，并更新 SNAPSHOT。
