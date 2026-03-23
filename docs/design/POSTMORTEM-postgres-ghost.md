# 心累史 — 幽灵 Postgres 事件

**日期**：2026-03-23
**历时**：约 3 小时
**结论**：macOS 本地有一个幽灵 postgres 实例在 127.0.0.1:5432，与 Docker 容器端口映射冲突，导致 Go 服务连错数据库，长达数小时的调试全部白费。

---

## 时间线

### 第一阶段：schema 改造
users 表加 email 字段，sqlc 重新生成，handler 改为用 email 登录注册。改动本身是正确的。

### 第二阶段：第一个坑——docker exec 和 Go 连的不是同一个库
`docker exec psql` 连的是 Docker 容器（正确）。
Go 程序连的是 `localhost:5432`（幽灵 postgres）。
两边数据完全不同，但没有人知道。

症状：
- docker exec 查 users 表是空的
- Go 服务说有 4 条 user 记录
- 注册报 username 唯一约束冲突，但 docker exec 查表是空的
- TRUNCATE 了，docker exec 查是 0，Go 还是说有数据

排查路径（全部白费）：
- 检查 SQL 语句 ✗
- 检查 sqlc 生成代码 ✗
- 检查 connect.go 顺序 ✗
- 重建 docker volume ✗（OrbStack volume 删不干净）
- 检查触发器 ✗
- 检查 etcd ✗
- 开启 postgres 慢查询日志 ✗
- REINDEX + VACUUM ✗
- 检查 pg_constraint ✗
- 加 debug log 查 users 表（发现 Go 能看到数据但 psql 看不到）

### 第三阶段：第二个坑——goreman 同时跑了两个 wallet-service
用 goreman 起了 etcd，但 goreman 的残留进程同时跑了 wallet-service。
加上手动 `go run` 的那个，同时有两个服务抢着处理同一个注册请求。
第一个服务先插入成功，第二个服务报约束冲突。

症状：注册偶发性成功，大部分时候失败。

### 第四阶段：第三个坑——OrbStack volume 删不干净
`docker volume rm` 显示成功，但重建容器后数据还在。
OrbStack 把 volume 数据持久化在独立位置，标准 docker 命令删不掉。

症状：明明删了 volume，数据依然存在。

### 第五阶段：破案
在 Go 的 connect.go 里加 debug log，打印实际连接的数据库名——是 `blitz`，没问题。
但通过 `docker inspect` 拿到容器 IP `192.168.117.2`，把 Go 的默认 DSN 改为这个 IP 后，注册立刻成功。

**根本原因**：
macOS 上有一个本地安装的 postgres（可能是 Homebrew 或其他工具附带），监听在 `127.0.0.1:5432`。
Docker 容器把端口映射到 `0.0.0.0:5432`，但在 macOS 上 `localhost` 解析为 `127.0.0.1`，被本地 postgres 截获了。
`lsof -i :5432` 显示 Docker 在监听，但实际上本地 postgres 在 `127.0.0.1` 上优先响应。

---

## 解决方案

临时方案（已应用）：
```go
dsn = "postgres://blitz:blitz@192.168.117.2:5432/blitz?sslmode=disable"
```

永久方案：
```bash
# 设置环境变量，优先级高于默认值
export DATABASE_URL="postgres://blitz:blitz@192.168.117.2:5432/blitz?sslmode=disable"
# 加到 .zshrc
```

根本解决（TODO）：
找到并停掉本地幽灵 postgres：
```bash
brew services list | grep postgresql
# 或
launchctl list | grep postgresql
```

---

## 教训

1. **端口冲突是最难排查的问题之一**，因为连接看起来成功了，数据就是对不上。
2. **OrbStack 的 volume 行为和标准 Docker 不同**，删 volume 要在 OrbStack UI 里操作。
3. **goreman 启动的进程要显式关闭**，不然会有残留进程。
4. **调试数据库问题时，第一步应该是验证两端连的是否同一个实例**，而不是排查 SQL。
5. **`lsof -i :5432` 不足以确认连接目标**，需要用容器 IP 而非 localhost 来绕过本地端口占用。

---

## 彩蛋

调试过程中 Claude 莫名其妙说了好几次韩语。原因不明。
