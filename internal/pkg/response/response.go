package response

import (
	"encoding/json"
	"net/http"

	"github.com/Ixecd/blitz/internal/pkg/code"
)

// Response 统一 HTTP 响应格式
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func OK(w http.ResponseWriter, data interface{}) {
	write(w, http.StatusOK, Response{Code: 0, Message: "ok", Data: data})
}

func Fail(w http.ResponseWriter, err *code.Error) {
	status := http.StatusInternalServerError
	switch err.Code {
	case code.ErrInvalidArg:
		status = http.StatusBadRequest
	case code.ErrUnauthorized:
		status = http.StatusUnauthorized
	case code.ErrForbidden:
		status = http.StatusForbidden
	case code.ErrNotFound:
		status = http.StatusNotFound
	}
	write(w, status, Response{
		Code:    int(err.Code),
		Message: err.Code.Message(),
	})
}

func write(w http.ResponseWriter, status int, resp Response) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(resp)
}
