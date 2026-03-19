# 错误码规范

> web3-blitz 统一错误码体系设计说明。

---

## 设计目标

- 前端根据 code 字段判断具体错误类型，不依赖 message 字符串
- 错误码和 HTTP 状态码解耦（同一个 HTTP 400 可以有多种业务错误码）
- 通过 codegen 工具自动生成代码和文档，避免手动维护两份

---

## 分段设计

| 段 | 范围 | 说明 |
|----|------|------|
| 通用错误 | 100000 - 100999 | 参数错误、未授权、服务器内部错误等 |
| 用户相关 | 101000 - 101999 | 注册、登录、token 等用户体系错误 |
| 钱包相关 | 102000 - 102999 | 充值、提币、余额、限额等钱包错误 |

后续扩展直接加新段，互不影响：

```
103xxx → 风控相关（待实现）
104xxx → KYC 相关（待实现）
105xxx → 审计相关（待实现）
```

---

## 当前错误码列表

### 通用错误（100xxx）

| 错误码 | 名称 | HTTP | 说明 |
|--------|------|------|------|
| 100000 | ErrUnknown | 500 | 未知错误 |
| 100001 | ErrInvalidArg | 400 | 参数错误 |
| 100002 | ErrUnauthorized | 401 | 未授权 |
| 100003 | ErrForbidden | 403 | 无权限 |
| 100004 | ErrNotFound | 404 | 资源不存在 |
| 100005 | ErrInternal | 500 | 服务器内部错误 |
| 100006 | ErrDeadlineExceeded | 408 | 请求超时 |

### 用户相关（101xxx）

| 错误码 | 名称 | HTTP | 说明 |
|--------|------|------|------|
| 101000 | ErrUserAlreadyExists | 409 | 用户名已存在 |
| 101001 | ErrUserNotFound | 404 | 用户不存在 |
| 101002 | ErrUserPasswordWrong | 401 | 用户名或密码错误 |
| 101003 | ErrUserTokenInvalid | 401 | token 无效或已过期 |
| 101004 | ErrUserRefreshTokenInvalid | 401 | refresh token 无效或已过期 |

### 钱包相关（102xxx）

| 错误码 | 名称 | HTTP | 说明 |
|--------|------|------|------|
| 102000 | ErrWalletChainNotSupported | 400 | 不支持的链 |
| 102001 | ErrWalletAddressInvalid | 400 | 无效的钱包地址 |
| 102002 | ErrWalletInsufficientBalance | 400 | 余额不足 |
| 102003 | ErrWalletDailyLimitExceeded | 400 | 超出每日提币限额 |
| 102004 | ErrWalletDuplicateWithdraw | 429 | 重复提币请求 |
| 102005 | ErrWalletBroadcastFailed | 500 | 交易广播失败 |

---

## 响应格式

成功：

```json
{"data": {"tx_id": "c8d0ae07...", "amount": 0.05}}
```

失败：

```json
{"code": 102002, "message": "Insufficient balance"}
```

前端判断逻辑：有 `code` 字段则失败，有 `data` 字段则成功。

---

## 代码实现

`internal/pkg/code/code.go` 定义错误码，注释格式严格遵循 `名称 - HTTP状态码: 描述.`：

```go
type ErrorCode int

const (
    ErrUnknown      ErrorCode = iota + 100000 // ErrUnknown - 500: Internal server error.
    ErrInvalidArg                             // ErrInvalidArg - 400: Invalid argument.
)

const (
    ErrUserAlreadyExists ErrorCode = iota + 101000 // ErrUserAlreadyExists - 409: User already exists.
)
```

`internal/api/response.go` 提供三个响应函数：

```go
OK(w, data)               // 成功，{"data": data}
Fail(w, code.ErrXxx)      // 失败，使用预定义 message
FailMsg(w, code.ErrXxx, "自定义上下文信息")  // 失败，携带上下文
```

handler 使用示例：

```go
// 直接用预定义 message
Fail(w, code.ErrUserAlreadyExists)
// → {"code": 101000, "message": "User already exists"}

// 带上下文（余额不足时附带具体数字）
FailMsg(w, code.ErrWalletInsufficientBalance,
    fmt.Sprintf("余额不足: 可用 %.8f，请求 %.8f", available, req.Amount))
// → {"code": 102002, "message": "余额不足: 可用 0.35000000，请求 3.00000000"}
```

---

## 新增错误码步骤

1. 在 `code.go` 对应段追加常量：

```go
ErrWalletNewError // ErrWalletNewError - 400: New error description.
```

2. 在 `code_generated.go` 的 `Message()` 和 `HTTPStatus()` 补充处理。

3. 运行 codegen 重新生成文档：

```bash
make gen.errcode.doc
```

---

## codegen 工具

dev-toolkit 里的 `tools/codegen` 扫描 Go 源码 AST，提取 `ErrorCode` 类型常量及注释，生成代码和文档。

```bash
# 生成 Go 代码
codegen -type=ErrorCode internal/pkg/code

# 生成 Markdown 文档
codegen -type=ErrorCode -doc \
  -output docs/guide/zh-CN/api/error_code_generated.md \
  internal/pkg/code
```

注释格式不符合规范时工具会降级到 500，并打印警告，不会崩溃。
