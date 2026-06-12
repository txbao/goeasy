package httpx

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/txbao/goeasy/casbin"
	zerr "github.com/txbao/goeasy/errors"
	"github.com/txbao/goeasy/response"
)

// RequireCasbin 校验 RBAC；subject 默认取 jwt_subject，可通过 c.Get("casbin_subject") 覆盖。
func RequireCasbin(enforcer *casbin.Enforcer, obj, act string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if enforcer == nil {
			response.FailBiz(c, http.StatusServiceUnavailable, int(zerr.BizCodeServiceUnavail), "casbin not enabled")
			c.Abort()
			return
		}
		sub, _ := c.Get("casbin_subject")
		subject, _ := sub.(string)
		if subject == "" {
			if v, ok := c.Get("jwt_subject"); ok {
				subject, _ = v.(string)
			}
		}
		if subject == "" {
			response.FailBiz(c, http.StatusUnauthorized, int(zerr.BizCodeAuthMissing), "missing subject")
			c.Abort()
			return
		}
		ok, err := enforcer.Check(subject, obj, act)
		if err != nil {
			response.FailInternal(c, err)
			c.Abort()
			return
		}
		if !ok {
			response.FailBiz(c, http.StatusForbidden, int(zerr.BizCodeForbidden), "permission denied")
			c.Abort()
			return
		}
		c.Next()
	}
}
