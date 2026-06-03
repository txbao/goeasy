package httpx

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/txbao/goeasy/config"
	"github.com/txbao/goeasy/health"
	"github.com/txbao/goeasy/limiter"
	"github.com/txbao/goeasy/metrics"
	"github.com/txbao/goeasy/trace"
)

// Options httpx 引擎选项。
type Options struct {
	Config  *config.Config
	Limiter *limiter.Limiter
	Health  *health.Registry
}

// NewEngine 创建带默认中间件的 Gin 引擎（P0–P3）。
func NewEngine(cfg *config.Config) *gin.Engine {
	return NewEngineWith(Options{Config: cfg})
}

// NewEngineWith 可注入限流与健康注册表。
func NewEngineWith(opt Options) *gin.Engine {
	cfg := opt.Config
	if cfg != nil && cfg.Env == "prod" {
		gin.SetMode(gin.ReleaseMode)
	}
	service := "goesy"
	if cfg != nil && cfg.AppName != "" {
		service = cfg.AppName
	}
	engine := gin.New()
	engine.Use(gin.Recovery())
	engine.Use(requestID())
	engine.Use(traceMiddleware())
	if opt.Limiter != nil {
		engine.Use(opt.Limiter.GinMiddleware())
	} else if cfg != nil && cfg.Governance.Limiter.Enabled {
		engine.Use(limiter.New(cfg.Governance.Limiter).GinMiddleware())
	}
	if cfg != nil && cfg.Observability.Metrics.Enabled {
		metrics.Init(cfg.Observability.Metrics, service)
		engine.Use(metrics.GinMiddleware(service))
		metrics.RegisterRoute(engine, cfg.Observability.Metrics.Path)
	}
	engine.Use(accessLog(cfg))
	if cfg != nil && cfg.Observability.Health.Enabled {
		health.RegisterRoutes(engine, cfg.Observability.Health, opt.Health)
	}
	return engine
}

func requestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.GetHeader("X-Request-ID")
		if id == "" {
			id = uuid.NewString()
		}
		c.Set("request_id", id)
		c.Header("X-Request-ID", id)
		c.Next()
	}
}

func traceMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, end := trace.Start(c.Request.Context(), c.FullPath())
		defer end()
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

func accessLog(cfg *config.Config) gin.HandlerFunc {
	service := "goesy"
	if cfg != nil && cfg.AppName != "" {
		service = cfg.AppName
	}
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		log.Printf(`{"service":%q,"method":%q,"path":%q,"status":%d,"latency_ms":%d,"request_id":%q,"trace_id":%q}`,
			service, c.Request.Method, c.Request.URL.Path, c.Writer.Status(),
			time.Since(start).Milliseconds(), c.GetString("request_id"),
			trace.TraceID(c.Request.Context()))
	}
}
