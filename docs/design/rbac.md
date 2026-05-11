# RBAC 权限系统设计

> web3-blitz 的基于角色的访问控制（Role-Based Access Control）设计文档。

---

## 为什么需要 RBAC

交易所钱包系统涉及资金安全，不同操作需要不同级别的权限控制：

- 普通用户只能操作自己的充提币
- 运营人员可以查看用户列表和限额配置，但不能修改
- 管理员可以修改限额、升级用户等级

简单的 `is_admin` 字段无法满足这种细粒度需求，RBAC 是行业标准解法。

---

## 数据模型

```
users ──── user_roles ──── roles ──── role_permissions ──── permissions
  │                          │                                    │
  id                         name(admin/operator/user)            name(user:read...)
  username                   description                          description
  level
```

四张表：

```sql
-- 角色表：定义系统中存在的角色
roles (id, name UNIQUE, description)

-- 权限表：定义系统中存在的权限点
permissions (id, name UNIQUE, description)

-- 角色-权限关联：多对多
role_permissions (role_id FK, permission_id FK, PRIMARY KEY(role_id, permission_id))

-- 用户-角色关联：多对多（一个用户可以有多个角色）
user_roles (user_id FK, role_id FK, PRIMARY KEY(user_id, role_id))
```

---

## 内置角色与权限

### 权限点

| 权限名 | 说明 |
|--------|------|
| user:read | 查看用户列表和详情 |
| user:upgrade | 升级用户等级 |
| limit:read | 查看提币限额配置 |
| limit:write | 修改提币限额配置 |

### 角色权限矩阵

| 角色 | user:read | user:upgrade | limit:read | limit:write |
|------|-----------|--------------|------------|-------------|
| admin | 是 | 是 | 是 | 是 |
| operator | 是 | 否 | 是 | 否 |
| user | 否 | 否 | 否 | 否 |

---

## 鉴权流程

```
HTTP 请求
    │
    ▼
JWTMiddleware
    ├─ 无 token → 401
    ├─ token 无效 → 401
    └─ token 有效 → 注入 claims 到 context
            │
            ▼
        RBACMiddleware(permission)
            ├─ 从 context 取 claims.UserID
            ├─ GetUserPermissions（JOIN 三张表）
            ├─ 检查是否包含所需权限
            ├─ 无权限 → 403
            └─ 有权限 → 执行 handler
```

权限查询 SQL（一次 JOIN 搞定）：

```sql
SELECT DISTINCT p.name FROM permissions p
JOIN role_permissions rp ON rp.permission_id = p.id
JOIN user_roles ur ON ur.role_id = rp.role_id
WHERE ur.user_id = $1;
```

---

## 中间件使用

```go
// 只需要 JWT（登录即可访问）
mux.HandleFunc("/api/v1/users/me",
    auth.JWTMiddleware(jwtSecret, h.GetMe))

// 需要 JWT + 特定权限（链式组合）
mux.HandleFunc("/api/v1/users",
    auth.JWTMiddleware(jwtSecret,
        auth.RBACMiddleware(queries, "user:read", h.ListUsers)))

mux.HandleFunc("/api/v1/withdrawal-limits/update",
    auth.JWTMiddleware(jwtSecret,
        auth.RBACMiddleware(queries, "limit:write", h.UpdateWithdrawalLimit)))
```

---

## 角色分配

当前方式（SQL）：

```sql
-- 分配角色
INSERT INTO user_roles (user_id, role_id)
SELECT u.id, r.id FROM users u, roles r
WHERE u.username = 'alice' AND r.name = 'admin'
ON CONFLICT DO NOTHING;

-- 查看用户角色
SELECT r.name FROM roles r
JOIN user_roles ur ON ur.role_id = r.id
WHERE ur.user_id = 1;

-- 撤销角色
DELETE FROM user_roles
WHERE user_id = 1
  AND role_id = (SELECT id FROM roles WHERE name = 'operator');
```

待实现接口：

```
POST   /api/v1/users/:id/roles  → 分配角色（需要 admin）
DELETE /api/v1/users/:id/roles  → 撤销角色（需要 admin）
```

---

## 扩展方式

新增权限点和角色完全是数据驱动，不需要修改任何代码：

```sql
-- 新增权限点
INSERT INTO permissions (name, description)
VALUES ('withdraw:audit', '提币审核权限');

-- 新增角色
INSERT INTO roles (name, description)
VALUES ('auditor', '审计人员');

-- 绑定权限
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p
WHERE r.name = 'auditor' AND p.name IN ('user:read', 'withdraw:audit');
```

---

## web3-blitz 集成

web3-blitz 骨架中 `internal/auth/rbac.go` 使用接口设计，不依赖具体 DB：

```go
type PermissionChecker interface {
    HasPermission(ctx context.Context, userID int64, permission string) (bool, error)
}
```

业务层实现接口，自由选择权限存储方式（DB、缓存、配置文件）。

---

## 待完善

- 角色分配/撤销接口
- 权限缓存（避免每次请求都 JOIN 三张表）
- 审计日志（记录谁在什么时候做了什么）
- 超级管理员保护（防止最后一个 admin 被降级）
