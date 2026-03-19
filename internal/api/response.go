package api

import (
	"encoding/json"
	"net/http"

	"github.com/Ixecd/web3-blitz/internal/pkg/code"
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

// FailMsg 带自定义 message 的错误响应
func FailMsg(w http.ResponseWriter, err code.ErrorCode, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(err.HTTPStatus())
	json.NewEncoder(w).Encode(Response{
		Code:    int(err),
		Message: msg,
	})
}