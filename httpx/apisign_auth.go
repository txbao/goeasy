package httpx

import (
	"bytes"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/txbao/goeasy/apisign"
	zerr "github.com/txbao/goeasy/errors"
	"github.com/txbao/goeasy/response"
)

// RequireAPISign 开放平台 RSA2 验签中间件。
func RequireAPISign(verifier *apisign.Verifier) gin.HandlerFunc {
	return func(c *gin.Context) {
		if verifier == nil {
			response.FailBiz(c, http.StatusServiceUnavailable, int(zerr.BizCodeServiceUnavail), "api_sign not enabled")
			c.Abort()
			return
		}
		var body []byte
		if c.Request.Body != nil {
			b, err := io.ReadAll(c.Request.Body)
			if err != nil {
				response.FailBiz(c, http.StatusBadRequest, int(zerr.BizCodeParamInvalid), "read body failed")
				c.Abort()
				return
			}
			body = b
			c.Request.Body = io.NopCloser(bytes.NewReader(body))
		}
		if err := verifier.VerifyRequest(c.Request, body); err != nil {
			response.FailBiz(c, http.StatusUnauthorized, int(zerr.BizCodeAuthMissing), err.Error())
			c.Abort()
			return
		}
		c.Next()
	}
}
