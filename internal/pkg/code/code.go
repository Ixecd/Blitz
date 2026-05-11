package code

		import "fmt"
		
		// ErrorCode 业务错误码
		type ErrorCode int
		
		const (
			ErrUnknown      ErrorCode = iota + 100000 // ErrUnknown - 500: Unknown error.
			ErrInvalidArg                             // ErrInvalidArg - 400: Invalid argument.
			ErrUnauthorized                           // ErrUnauthorized - 401: Unauthorized.
			ErrForbidden                              // ErrForbidden - 403: Forbidden.
			ErrNotFound                               // ErrNotFound - 404: Not found.
			ErrInternal                               // ErrInternal - 500: Internal server error.
		)
		
		var errorCodeMessages = map[ErrorCode]string{
			ErrUnknown:      "unknown error",
			ErrInvalidArg:   "invalid argument",
			ErrUnauthorized: "unauthorized",
			ErrForbidden:    "forbidden",
			ErrNotFound:     "not found",
			ErrInternal:     "internal server error",
		}
		
		func (e ErrorCode) Message() string {
			if msg, ok := errorCodeMessages[e]; ok {
				return msg
			}
			return fmt.Sprintf("error code %d", int(e))
		}
		
		// Error 业务错误，携带错误码 + 原因
		type Error struct {
			Code  ErrorCode
			Cause error
		}
		
		func New(code ErrorCode) *Error { return &Error{Code: code} }
		
		func (e *Error) WithCause(cause error) *Error {
			e.Cause = cause
			return e
		}
		
		func (e *Error) Error() string {
			if e.Cause != nil {
				return fmt.Sprintf("[%d] %s: %v", e.Code, e.Code.Message(), e.Cause)
			}
			return fmt.Sprintf("[%d] %s", e.Code, e.Code.Message())
		}
		