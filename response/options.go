package response

import (
	"sync"

	"github.com/txbao/goeasy/logger"
)

var (
	optMu    sync.RWMutex
	settings = struct {
		logger          *logger.Logger
		env             string
		logServerErrors bool
		exposeDetail    *bool
	}{
		logServerErrors: true,
	}
)

// Options HTTP 错误响应与日志行为（由 httpx.NewEngineWith 注入）。
type Options struct {
	Logger            *logger.Logger
	Env               string
	LogServerErrors   bool
	ExposeErrorDetail *bool // nil 表示按 env：dev 暴露、prod 脱敏
}

// Configure 设置包级 logger、环境与 HTTP 观测选项。
func Configure(o Options) {
	optMu.Lock()
	defer optMu.Unlock()
	if o.Logger != nil {
		settings.logger = o.Logger
	}
	if o.Env != "" {
		settings.env = o.Env
	}
	settings.logServerErrors = o.LogServerErrors
	if o.ExposeErrorDetail != nil {
		settings.exposeDetail = o.ExposeErrorDetail
	}
}

// SetLogger 注入结构化 logger（供 httpx 启动时调用）。
func SetLogger(log *logger.Logger) {
	optMu.Lock()
	defer optMu.Unlock()
	settings.logger = log
}

// SetEnv 设置运行环境（dev/prod），控制 FailInternal 响应脱敏。
func SetEnv(env string) {
	optMu.Lock()
	defer optMu.Unlock()
	settings.env = env
}

func snapshot() (log *logger.Logger, env string, logServerErrors bool, exposeDetail *bool) {
	optMu.RLock()
	defer optMu.RUnlock()
	return settings.logger, settings.env, settings.logServerErrors, settings.exposeDetail
}

func shouldExposeDetail(env string, override *bool) bool {
	if override != nil {
		return *override
	}
	return env == "dev"
}
