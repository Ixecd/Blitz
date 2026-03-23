# SNAPSHOT — web3-blitz

**里程碑**：前后端登录打通 + 管理后台可用
**日期**：2026-03-23

---

## 已完成

### 后端
- users 表加 email 字段（migration + schema.sql 同步更新）
- Login / Register 接口从 username 改为 email
- sqlc 新增 GetUserByEmail 查询
- connect.go：NewDB 去掉 path 参数，从环境变量读 DATABASE_URL
- connect.go：启动时自动执行 ALTER TABLE 加 email 字段（兼容旧库）
- roles / permissions / role_permissions seed 数据写入 schema.sql
- 发现并修复：macOS localhost:5432 被本地幽灵 postgres 占用，默认 DSN 改为容器 IP

### 前端
- AuthContext：username 全面改为 email，接口 body 改为 { email, password }
- api/client.ts：修复 json.code !== undefined 判断，解决 200 OK 被误判为错误
- Layout.tsx：显示 email 而非 username，去掉「已登录」文字，Logo 放大
- Login.tsx：非受控输入（useRef），解决 autofill 显示问题
- index.css：data-table th 颜色从 text-faint 改为 text-muted

### 基础设施
- 发现 goreman 和手动 go run 同时跑两个 wallet-service 导致注册冲突
- 发现 OrbStack volume 删除不彻底，数据持久化在独立位置
- 确认 macOS 本地有幽灵 postgres，与 Docker 容器端口冲突

---

## 当前可用账号
- superadmin@blitz.com / admin123456（admin 角色）

## 环境变量（必须设置）
```bash
export DATABASE_URL="postgres://blitz:blitz@192.168.117.2:5432/blitz?sslmode=disable"
```

---

## 待完成
- 忘记密码 / 注册页面 / 扫码登录
- Dashboard / Deposit / Withdraw 页面按新设计语言重做
- 记住我 cookie 逻辑
- connect.go 中临时 ALTER TABLE migration 代码清理
- 解决 macOS 本地幽灵 postgres 根本原因
