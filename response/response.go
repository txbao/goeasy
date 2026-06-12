package response

import (
	"encoding/json"
	"net/http"
)

// Response 标准 JSON 响应体（与 Gin body 一致）。
type Response struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data any    `json:"data,omitempty"`
}

func JSON(w http.ResponseWriter, code int, msg string, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	resp := Response{
		Code: code,
		Msg:  msg,
		Data: data,
	}
	json.NewEncoder(w).Encode(resp)
}

// JSONBiz 非 Gin 场景：HTTP Status 与 body.code 分离。
func JSONBiz(w http.ResponseWriter, httpCode int, bizCode int, msg string, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpCode)
	resp := Response{
		Code: bizCode,
		Msg:  msg,
		Data: data,
	}
	json.NewEncoder(w).Encode(resp)
}
