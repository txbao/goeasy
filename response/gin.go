package response

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	zerr "github.com/txbao/goeasy/errors"
)

type body struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data any    `json:"data,omitempty"`
}

// Success Gin 统一成功响应。
func Success(c *gin.Context, data any) {
	c.JSON(http.StatusOK, body{Code: 0, Msg: "success", Data: data})
}

// Fail Gin 统一失败响应。
func Fail(c *gin.Context, httpCode int, msg string) {
	if httpCode <= 0 {
		httpCode = http.StatusInternalServerError
	}
	c.JSON(httpCode, body{Code: httpCode, Msg: msg})
}

// FailErr 从 error 推断响应（支持 CodedError）。
func FailErr(c *gin.Context, err error) {
	if err == nil {
		Success(c, nil)
		return
	}
	var ce *zerr.CodedError
	if errors.As(err, &ce) {
		httpCode := http.StatusInternalServerError
		switch ce.Code {
		case zerr.CodeNotFound:
			httpCode = http.StatusNotFound
		case zerr.CodeInvalid:
			httpCode = http.StatusBadRequest
		case zerr.CodeForbidden:
			httpCode = http.StatusForbidden
		case zerr.CodeUnauthorized:
			httpCode = http.StatusUnauthorized
		}
		Fail(c, httpCode, ce.Msg)
		return
	}
	Fail(c, http.StatusInternalServerError, err.Error())
}
