package api

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/Ixecd/blitz/internal/code"
)

type Response struct {
	Code    int         `json:"code,omitempty"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

// OK 成功响应
func OK(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(Response{Data: data})
}

// Fail 错误响应
func Fail(w http.ResponseWriter, err code.ErrorCode) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(err.HTTPStatus())
	json.NewEncoder(w).Encode(Response{
		Code:    int(err),
		Message: err.Message(),
	})
}

// FailMsg 带自定义 message 的错误响应（用于业务逻辑错误）
func FailMsg(w http.ResponseWriter, err code.ErrorCode, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(err.HTTPStatus())
	json.NewEncoder(w).Encode(Response{
		Code:    int(err),
		Message: msg,
	})
}

// FailInternal 内部错误 — 记录真实错误，向用户返回通用消息
func FailInternal(w http.ResponseWriter, internalErr error) {
	slog.Error("internal error", "err", internalErr)
	w.Header().Set("Content-Type", "application/json")
	ec := code.ErrInternal
	w.WriteHeader(ec.HTTPStatus())
	json.NewEncoder(w).Encode(Response{
		Code:    int(ec),
		Message: "服务内部错误，请稍后重试",
	})
}