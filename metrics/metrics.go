package metrics

import (
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/txbao/goeasy/config"
)

var (
	once      sync.Once
	httpTotal *prometheus.CounterVec
	httpDur   *prometheus.HistogramVec
	enabled   bool
)

// Init 注册 Prometheus 指标。
func Init(cfg config.MetricsCfg, service string) {
	if !cfg.Enabled {
		return
	}
	once.Do(func() {
		httpTotal = prometheus.NewCounterVec(
			prometheus.CounterOpts{Name: "goesy_http_requests_total", Help: "HTTP requests"},
			[]string{"service", "method", "path", "status"},
		)
		httpDur = prometheus.NewHistogramVec(
			prometheus.HistogramOpts{Name: "goesy_http_duration_ms", Help: "HTTP latency ms"},
			[]string{"service", "method", "path"},
		)
		prometheus.MustRegister(httpTotal, httpDur)
		enabled = true
	})
	_ = service
}

// GinMiddleware 采集 HTTP 指标。
func GinMiddleware(service string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !enabled {
			c.Next()
			return
		}
		start := time.Now()
		c.Next()
		path := c.FullPath()
		if path == "" {
			path = c.Request.URL.Path
		}
		status := c.Writer.Status()
		ms := float64(time.Since(start).Milliseconds())
		httpTotal.WithLabelValues(service, c.Request.Method, path, strconv.Itoa(status)).Inc()
		httpDur.WithLabelValues(service, c.Request.Method, path).Observe(ms)
	}
}

// RegisterRoute 暴露 /metrics。
func RegisterRoute(r *gin.Engine, path string) {
	if !enabled {
		return
	}
	if path == "" {
		path = "/metrics"
	}
	r.GET(path, gin.WrapH(promhttp.Handler()))
}
