package app

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/txbao/goeasy/audit"
	"github.com/txbao/goeasy/breaker"
	"github.com/txbao/goeasy/cache"
	"github.com/txbao/goeasy/casbin"
	"github.com/txbao/goeasy/config"
	"github.com/txbao/goeasy/database"
	"github.com/txbao/goeasy/discovery"
	"github.com/txbao/goeasy/grpcx"
	"github.com/txbao/goeasy/health"
	"github.com/txbao/goeasy/httpx"
	"github.com/txbao/goeasy/idgen"
	"github.com/txbao/goeasy/jwt"
	"github.com/txbao/goeasy/limiter"
	"github.com/txbao/goeasy/logger"
	"github.com/txbao/goeasy/metrics"
	"github.com/txbao/goeasy/mq"
	"github.com/txbao/goeasy/scheduler"
	"github.com/txbao/goeasy/storage"
	"github.com/txbao/goeasy/trace"
)

// HTTPRegister 业务 HTTP 路由注册函数。
type HTTPRegister func(*gin.Engine)

// App 应用生命周期（统一装配 HTTP 与 P0–P4 组件）。
type App struct {
	cfg     *config.Config
	log     *logger.Logger
	engine  *gin.Engine
	httpReg HTTPRegister
	httpSrv *http.Server

	// P1
	DB       database.DB
	Cache    cache.Cache
	MQ       mq.MQ
	GRPC     *grpcx.Server
	Registry discovery.Registry
	Store    storage.Store
	Cron     scheduler.Scheduler

	// P2
	CircuitBreaker *breaker.Breaker
	Limiter        *limiter.Limiter

	// P3
	HealthReg *health.Registry
	Audit     *audit.Logger

	// P4
	IDGen  *idgen.Generator
	JWT    *jwt.Token
	Casbin *casbin.Enforcer
}

func New(cfg *config.Config) *App {
	if cfg == nil {
		panic("goeasy app: config is nil")
	}
	a := &App{
		cfg:            cfg,
		log:            logger.New(cfg),
		CircuitBreaker: breaker.New(cfg.Governance.Breaker),
		Limiter:        limiter.New(cfg.Governance.Limiter),
		HealthReg:      health.NewRegistry(),
		Audit:          audit.New(cfg.Observability.Audit),
		IDGen:          idgen.New(cfg.Enterprise.IDGen),
	}
	trace.Init(cfg.Observability.Trace)
	metrics.Init(cfg.Observability.Metrics, cfg.AppName)
	var err error
	a.JWT, err = jwt.New(cfg.Enterprise.JWT)
	if err != nil {
		panic(fmt.Sprintf("goeasy jwt: %v", err))
	}
	a.Casbin, err = casbin.New(cfg.Enterprise.Casbin)
	if err != nil {
		panic(fmt.Sprintf("goeasy casbin: %v", err))
	}
	a.engine = httpx.NewEngineWith(httpx.Options{
		Config:  cfg,
		Limiter: a.Limiter,
		Health:  a.HealthReg,
	})
	return a
}

func (a *App) RegisterHTTP(reg HTTPRegister) *App {
	a.httpReg = reg
	return a
}

// RegisterHealth 注册基础设施健康检查项。
func (a *App) RegisterHealth(c health.Checker) *App {
	a.HealthReg.Register(c)
	return a
}

func (a *App) InitInfra() error {
	var err error
	if a.cfg.DB.Enabled {
		a.DB, err = database.Open(a.cfg.DB)
		if err != nil {
			return fmt.Errorf("database: %w", err)
		}
	} else {
		a.DB = database.NewNoop()
	}
	if a.cfg.Redis.Enabled {
		a.Cache, err = cache.Open(a.cfg.Redis)
		if err != nil {
			return fmt.Errorf("cache: %w", err)
		}
	} else {
		a.Cache = cache.NewNoop()
	}
	if a.cfg.MQ.Enabled {
		a.MQ, err = mq.Open(a.cfg.MQ)
		if err != nil {
			return fmt.Errorf("mq: %w", err)
		}
	} else {
		a.MQ = mq.NewNoop()
	}
	if a.cfg.GRPC.Enabled {
		a.GRPC, err = grpcx.NewServer(a.cfg.GRPC)
		if err != nil {
			return fmt.Errorf("grpc: %w", err)
		}
	}
	a.Registry = discovery.NewRegistry(a.cfg)
	a.Store = storage.NewStore(a.cfg)
	a.Cron = scheduler.NewScheduler(a.cfg)
	return nil
}

func (a *App) Config() *config.Config    { return a.cfg }
func (a *App) Logger() *logger.Logger    { return a.log }
func (a *App) Engine() *gin.Engine       { return a.engine }
func (a *App) Breaker() *breaker.Breaker { return a.CircuitBreaker }

// NewGRPCClient 创建带治理的 gRPC 客户端。
func (a *App) NewGRPCClient(target string) (*grpcx.Client, error) {
	return grpcx.NewClient(grpcx.ClientOption{
		Target:     target,
		TimeoutSec: a.cfg.GRPC.TimeoutSec,
		MaxRetries: a.cfg.GRPC.MaxRetries,
		Cfg:        a.cfg,
		Strategy:   "round_robin",
	})
}

func (a *App) Run() error {
	if err := a.InitInfra(); err != nil {
		return err
	}
	if a.httpReg != nil {
		a.httpReg(a.engine)
	}
	a.httpSrv = &http.Server{
		Addr:    a.cfg.HTTPAddr(),
		Handler: a.engine,
	}
	errCh := make(chan error, 1)
	go func() {
		a.log.Infof("http listening on %s", a.cfg.HTTPAddr())
		if err := a.httpSrv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	select {
	case err := <-errCh:
		return err
	case <-quit:
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	a.log.Infof("shutting down...")
	return a.httpSrv.Shutdown(ctx)
}
