package health

import (
	"context"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"

	"github.com/txbao/goeasy/config"
	zresp "github.com/txbao/goeasy/response"
)

// Checker 健康检查项。
type Checker interface {
	Name() string
	Check(ctx context.Context) error
}

// Registry 探针注册表。
type Registry struct {
	mu       sync.RWMutex
	checkers []Checker
}

func NewRegistry() *Registry {
	return &Registry{}
}

func (r *Registry) Register(c Checker) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.checkers = append(r.checkers, c)
}

func (r *Registry) CheckAll(ctx context.Context) map[string]string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make(map[string]string, len(r.checkers)+1)
	out["framework"] = "ok"
	for _, c := range r.checkers {
		if err := c.Check(ctx); err != nil {
			out[c.Name()] = err.Error()
		} else {
			out[c.Name()] = "ok"
		}
	}
	return out
}

// RegisterRoutes 注册框架健康探针（与业务 /health 并存）。
func RegisterRoutes(r *gin.Engine, cfg config.HealthCfg, reg *Registry) {
	if !cfg.Enabled {
		return
	}
	path := cfg.Path
	if path == "" {
		path = "/healthz"
	}
	if reg == nil {
		reg = NewRegistry()
	}
	r.GET(path, func(c *gin.Context) {
		status := reg.CheckAll(c.Request.Context())
		zresp.Success(c, gin.H{"checks": status})
	})
}

// PingHandler 简单存活探针。
func PingHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "up"})
	}
}
