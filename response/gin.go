package response

import (
	stderrors "errors"
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
// Deprecated: 新业务请使用 FailBiz；body.code 与 httpCode 相同，不符合 6 位业务码规范。
func Fail(c *gin.Context, httpCode int, msg string) {
	if httpCode <= 0 {
		httpCode = http.StatusInternalServerError
	}
	logServerError(c, httpCode, httpCode, nil, msg)
	c.JSON(httpCode, body{Code: httpCode, Msg: msg})
}

// FailBiz 标准业务错误响应：HTTP Status 与 body.code（6 位业务码）分离。
func FailBiz(c *gin.Context, httpCode int, bizCode int, msg string) {
	if httpCode <= 0 {
		httpCode = http.StatusInternalServerError
	}
	logServerError(c, httpCode, bizCode, nil, msg)
	c.JSON(httpCode, body{Code: bizCode, Msg: msg})
}

// FailInternal 未映射的内部错误兜底（bizCode 固定 500001）；prod 响应脱敏，日志保留完整 err。
func FailInternal(c *gin.Context, err error) {
	_, env, _, exposeOverride := snapshot()
	msg := "服务内部错误"
	if shouldExposeDetail(env, exposeOverride) && err != nil {
		msg = err.Error()
	}
	logServerError(c, http.StatusInternalServerError, int(zerr.BizCodeInternal), err, msg)
	c.JSON(http.StatusInternalServerError, body{
		Code: int(zerr.BizCodeInternal),
		Msg:  msg,
	})
}

// FailErr 从 error 推断响应：优先 BizError，其次 CodedError，最后 FailInternal。
func FailErr(c *gin.Context, err error) {
	if err == nil {
		Success(c, nil)
		return
	}
	var be *zerr.BizError
	if stderrors.As(err, &be) {
		httpCode := httpCodeForBiz(be.Code)
		FailBiz(c, httpCode, int(be.Code), be.Msg)
		return
	}
	var ce *zerr.CodedError
	if stderrors.As(err, &ce) {
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
	FailInternal(c, err)
}

func httpCodeForBiz(code zerr.BizCode) int {
	switch code {
	case zerr.BizCodeParamInvalid, zerr.BizCodeParamFormat:
		return http.StatusBadRequest
	case zerr.BizCodeUnauthorized, zerr.BizCodeAuthMissing:
		return http.StatusUnauthorized
	case zerr.BizCodeForbidden:
		return http.StatusForbidden
	case zerr.BizCodeServiceUnavail:
		return http.StatusServiceUnavailable
	default:
		if int(code) >= 500000 {
			return http.StatusInternalServerError
		}
		return http.StatusBadRequest
	}
}
