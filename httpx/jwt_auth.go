package httpx

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/txbao/goeasy/jwt"
	"github.com/txbao/goeasy/response"
)

// RequireJWT 校验 Bearer JWT；token 为 nil 时返回 503（未启用）。
func RequireJWT(token *jwt.Token, header string) gin.HandlerFunc {
	if header == "" {
		header = "Authorization"
	}
	return func(c *gin.Context) {
		if token == nil {
			response.Fail(c, http.StatusServiceUnavailable, "jwt not enabled")
			c.Abort()
			return
		}
		raw := c.GetHeader(header)
		if raw == "" {
			response.Fail(c, http.StatusUnauthorized, "missing authorization")
			c.Abort()
			return
		}
		if strings.HasPrefix(strings.ToLower(raw), "bearer ") {
			raw = strings.TrimSpace(raw[7:])
		}
		claims, err := token.Parse(raw)
		if err != nil {
			response.Fail(c, http.StatusUnauthorized, "invalid token")
			c.Abort()
			return
		}
		c.Set("jwt_subject", claims.Subject)
		c.Set("jwt_claims", claims)
		c.Next()
	}
}
