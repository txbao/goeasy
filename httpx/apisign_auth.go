package httpx

import (
	"bytes"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/txbao/goeasy/apisign"
	"github.com/txbao/goeasy/response"
)

// RequireAPISign 开放平台 RSA2 验签中间件。
func RequireAPISign(verifier *apisign.Verifier) gin.HandlerFunc {
	return func(c *gin.Context) {
		if verifier == nil {
			response.Fail(c, http.StatusServiceUnavailable, "api_sign not enabled")
			c.Abort()
			return
		}
		var body []byte
		if c.Request.Body != nil {
			b, err := io.ReadAll(c.Request.Body)
			if err != nil {
				response.Fail(c, http.StatusBadRequest, "read body failed")
				c.Abort()
				return
			}
			body = b
			c.Request.Body = io.NopCloser(bytes.NewReader(body))
		}
		if err := verifier.VerifyRequest(c.Request, body); err != nil {
			response.Fail(c, http.StatusUnauthorized, err.Error())
			c.Abort()
			return
		}
		c.Next()
	}
}
