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
	once        sync.Once
	httpTotal   *prometheus.CounterVec
	httpDur     *prometheus.HistogramVec
	sqlTotal    *prometheus.CounterVec
	sqlSlow     *prometheus.CounterVec
	cacheHit    *prometheus.CounterVec
	cacheMiss   *prometheus.CounterVec
	enabled     bool
	serviceName string
)

// Init 注册 Prometheus 指标。
func Init(cfg config.MetricsCfg, service string) {
	if !cfg.Enabled {
		return
	}
	once.Do(func() {
		serviceName = service
		httpTotal = prometheus.NewCounterVec(
			prometheus.CounterOpts{Name: "goeasy_http_requests_total", Help: "HTTP requests"},
			[]string{"service", "method", "path", "status"},
		)
		httpDur = prometheus.NewHistogramVec(
			prometheus.HistogramOpts{Name: "goeasy_http_duration_ms", Help: "HTTP latency ms"},
			[]string{"service", "method", "path"},
		)
		sqlTotal = prometheus.NewCounterVec(
			prometheus.CounterOpts{Name: "goeasy_sql_queries_total", Help: "SQL queries"},
			[]string{"service", "op"},
		)
		sqlSlow = prometheus.NewCounterVec(
			prometheus.CounterOpts{Name: "goeasy_sql_slow_total", Help: "Slow SQL queries"},
			[]string{"service", "op"},
		)
		cacheHit = prometheus.NewCounterVec(
			prometheus.CounterOpts{Name: "goeasy_cache_hit_total", Help: "Cache hits"},
			[]string{"service", "layer"},
		)
		cacheMiss = prometheus.NewCounterVec(
			prometheus.CounterOpts{Name: "goeasy_cache_miss_total", Help: "Cache misses"},
			[]string{"service", "layer"},
		)
		prometheus.MustRegister(httpTotal, httpDur, sqlTotal, sqlSlow, cacheHit, cacheMiss)
		enabled = true
	})
	if service != "" {
		serviceName = service
	}
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

// ObserveSQL 记录 SQL 执行指标。
func ObserveSQL(op string, slow bool) {
	if !enabled {
		return
	}
	sqlTotal.WithLabelValues(serviceName, op).Inc()
	if slow {
		sqlSlow.WithLabelValues(serviceName, op).Inc()
	}
}

// ObserveCacheHit 缓存命中。
func ObserveCacheHit(layer string) {
	if !enabled {
		return
	}
	cacheHit.WithLabelValues(serviceName, layer).Inc()
}

// ObserveCacheMiss 缓存未命中。
func ObserveCacheMiss(layer string) {
	if !enabled {
		return
	}
	cacheMiss.WithLabelValues(serviceName, layer).Inc()
}
