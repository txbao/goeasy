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

	"github.com/txbao/goeasy/apisign"
	"github.com/txbao/goeasy/audit"
	"github.com/txbao/goeasy/breaker"
	"github.com/txbao/goeasy/cache"
	"github.com/txbao/goeasy/casbin"
	"github.com/txbao/goeasy/eventbus"
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
	"github.com/txbao/goeasy/outbox"
	"github.com/txbao/goeasy/scheduler"
	"github.com/txbao/goeasy/sqllog"
	"github.com/txbao/goeasy/storage"
	"github.com/txbao/goeasy/trace"
)

// HTTPInfra 注入到业务 bootstrap 的基础设施（在 InitInfra 之后可用）。
type HTTPInfra struct {
	DB          database.DB
	DBDriver    string // database.driver：postgres | mysql
	TablePrefix string // database.table_prefix

	Cache              cache.Cache      // redis.enabled 时为真实客户端，否则 Noop
	Locker             cache.Locker     // 分布式锁（redis.enabled）
	GuardedGet         *cache.GuardedGet // 防击穿 singleflight
	RedisKeyPrefix     string           // 实体缓存 key 前缀，见 config.RedisKeyPrefix()
	EntityCacheEnabled bool             // redis.enabled && cache.enabled
	EntityCacheTTL     time.Duration    // 实体缓存 TTL，默认一周
	NullCacheTTL       time.Duration    // 空值缓存 TTL

	JWT       *jwt.Token          // enterprise.jwt（管理后台）
	MemberJWT *jwt.Token          // enterprise.member_jwt（H5 会员）
	Casbin    *casbin.Enforcer    // enterprise.casbin（可选 RBAC）
	APISign   *apisign.Verifier   // enterprise.api_sign（开放平台验签）
	MQ        mq.MQ               // mq.enabled 时为真实客户端，否则 Noop
	Outbox    *outbox.Publisher   // mq.outbox 发布器（默认直发 MQ）
	MQLog     config.MQLogCfg     // observability.mq 日志选项
	EventBus  eventbus.Bus        // 进程内事件总线
	Cron      scheduler.Scheduler // 定时任务

	GRPCClientForService func(context.Context, string) (*grpcx.Client, error)
	RPC                  *grpcx.Registry // 长连接 RPC 客户端池（bootstrap / Gateway 使用）

	AuditRecorder audit.Recorder // 业务操作日志 Port（默认 NopRecorder）
}

// HTTPRegister 业务 HTTP 路由注册函数；返回 error 时 Run 中止启动（不监听 HTTP）。
type HTTPRegister func(*gin.Engine, HTTPInfra) error

// GRPCRegister 业务 gRPC 服务注册（在 Run 前调用，向 *grpcx.Server 注册 pb Service）。
type GRPCRegister func(*grpcx.Server, HTTPInfra)

// ConsumerRegister 业务 MQ 消费者注册（cmd/consumer 使用）。
type ConsumerRegister func(mq.MQ, HTTPInfra)

// App 应用生命周期（统一装配 HTTP 与 P0–P4 组件）。
type App struct {
	cfg     *config.Config
	log     *logger.Logger
	engine  *gin.Engine
	httpReg      HTTPRegister
	grpcReg      GRPCRegister
	consumerReg  ConsumerRegister
	httpSrv         *http.Server
	consumerHttpSrv *http.Server

	// P1
	DB       database.DB
	Cache    cache.Cache
	MQ       mq.MQ
	GRPC     *grpcx.Server
	Registry discovery.Registry
	GRPCRegistry *grpcx.Registry
	Store    storage.Store
	Cron     scheduler.Scheduler

	// P2
	CircuitBreaker *breaker.Breaker
	Limiter        *limiter.Limiter

	// P3
	HealthReg *health.Registry
	Audit     *audit.Logger

	// 业务操作日志 Recorder（默认 Nop；bootstrap 可通过 SetAuditRecorder 注入 PG 实现）
	auditRecorder         audit.Recorder
	auditRecorderRuntime  audit.Recorder

	// P4
	IDGen      *idgen.Generator
	JWT        *jwt.Token // 管理后台
	MemberJWT  *jwt.Token // H5 会员（member_jwt.enabled）
	Casbin     *casbin.Enforcer
	APISign    *apisign.Verifier
	EventBus   eventbus.Bus
	OutboxPub  *outbox.Publisher
	OutboxRelay *outbox.Relay
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
		auditRecorder:  audit.NopRecorder{},
		IDGen:          idgen.New(cfg.Enterprise.IDGen),
	}
	if err := trace.Init(cfg.Observability.Trace, cfg.AppName); err != nil {
		panic(fmt.Sprintf("goeasy trace: %v", err))
	}
	metrics.Init(cfg.Observability.Metrics, cfg.AppName)
	var err error
	a.JWT, err = jwt.New(cfg.Enterprise.JWT)
	if err != nil {
		panic(fmt.Sprintf("goeasy jwt: %v", err))
	}
	a.MemberJWT, err = jwt.New(cfg.Enterprise.MemberJWT)
	if err != nil {
		panic(fmt.Sprintf("goeasy member_jwt: %v", err))
	}
	a.Casbin, err = casbin.New(cfg.Enterprise.Casbin)
	if err != nil {
		panic(fmt.Sprintf("goeasy casbin: %v", err))
	}
	a.APISign, err = apisign.New(cfg.Enterprise.APISign)
	if err != nil {
		panic(fmt.Sprintf("goeasy api_sign: %v", err))
	}
	a.EventBus = eventbus.New()
	sqllog.Init(cfg.Observability.SQL)
	a.Cron = scheduler.NewScheduler(cfg.Scheduler)
	a.engine = httpx.NewEngineWith(httpx.Options{
		Config:  cfg,
		Logger:  a.log,
		Limiter: a.Limiter,
		Health:  a.HealthReg,
	})
	return a
}

func (a *App) RegisterHTTP(reg HTTPRegister) *App {
	a.httpReg = reg
	return a
}

// RegisterGRPC 注册 gRPC 服务实现（需 grpc.enabled）。
func (a *App) RegisterGRPC(reg GRPCRegister) *App {
	a.grpcReg = reg
	return a
}

// RegisterConsumer 注册 MQ 消费者（cmd/consumer 独立进程使用）。
func (a *App) RegisterConsumer(reg ConsumerRegister) *App {
	a.consumerReg = reg
	return a
}

// SetAuditRecorder 注入业务操作日志持久化实现（在 Run 前调用）。
func (a *App) SetAuditRecorder(r audit.Recorder) *App {
	if r != nil {
		a.auditRecorder = r
	}
	return a
}

// RegisterHealth 注册基础设施健康检查项。
func (a *App) RegisterHealth(c health.Checker) *App {
	a.HealthReg.Register(c)
	return a
}

// RegisterCron 注册定时任务（需在 Run 前调用；kind=system|business）。
func (a *App) RegisterCron(kind scheduler.TaskKind, name, spec string, fn func(context.Context) error) *App {
	if a.Cron != nil {
		_ = a.Cron.RegisterKind(kind, name, spec, fn)
	}
	return a
}

func (a *App) InitInfra() error {
	var err error
	if a.cfg.DB.Enabled {
		a.DB, err = database.Open(a.cfg.DB)
		if err != nil {
			return fmt.Errorf("database: %w", err)
		}
		if err := a.DB.Ping(context.Background()); err != nil {
			return fmt.Errorf("database ping: %w", err)
		}
	} else {
		a.DB = database.NewNoop()
	}
	if a.cfg.Redis.Enabled {
		base, err := cache.Open(a.cfg.Redis)
		if err != nil {
			return fmt.Errorf("cache: %w", err)
		}
		if a.cfg.Cache.L1.Enabled {
			l1 := cache.NewL1(a.cfg.Cache.L1.MaxEntries, a.cfg.L1CacheTTL())
			a.Cache = cache.NewMultiLevel(l1, base)
		} else {
			a.Cache = base
		}
		a.Limiter.BindRedis(cache.RedisClient(base))
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
	a.GRPCRegistry = grpcx.NewRegistry(a.cfg, a.Registry)
	a.Store = storage.NewStore(a.cfg)
	a.OutboxPub = outbox.NewPublisher(a.cfg, a.MQ, a.DB)
	if a.cfg.MQ.Outbox.Enabled {
		if !a.cfg.DB.Enabled {
			return fmt.Errorf("outbox requires database.enabled")
		}
		if !a.cfg.MQ.Enabled {
			return fmt.Errorf("outbox requires mq.enabled")
		}
		a.OutboxRelay = outbox.NewRelay(a.cfg, a.MQ, a.DB)
		if a.OutboxRelay == nil {
			return fmt.Errorf("outbox relay init failed")
		}
		spec := outbox.PollCronSpec(a.cfg.OutboxPollInterval())
		_ = a.RegisterCron(scheduler.TaskSystem, "outbox_relay", spec, a.OutboxRelay.RunOnce)
	}
	return nil
}

func (a *App) Config() *config.Config    { return a.cfg }
func (a *App) Logger() *logger.Logger    { return a.log }
func (a *App) Engine() *gin.Engine       { return a.engine }
func (a *App) Breaker() *breaker.Breaker { return a.CircuitBreaker }

// NewGRPCClient 使用显式 target（host:port）创建客户端。
func (a *App) NewGRPCClient(target string) (*grpcx.Client, error) {
	return grpcx.NewClient(grpcx.ClientOption{
		Target:     target,
		TimeoutSec: a.cfg.GRPC.TimeoutSec,
		MaxRetries: a.cfg.GRPC.MaxRetries,
		Cfg:        a.cfg,
		Strategy:   "round_robin",
	})
}

// NewGRPCClientForService 按逻辑服务名解析 target（discovery.services 或 etcd）。
func (a *App) NewGRPCClientForService(ctx context.Context, service string) (*grpcx.Client, error) {
	target, err := grpcx.ResolveService(ctx, a.cfg, a.Registry, service)
	if err != nil {
		return nil, err
	}
	return a.NewGRPCClient(target)
}

func (a *App) Run() error {
	if err := a.InitInfra(); err != nil {
		return err
	}
	infra := a.buildHTTPInfra()
	if a.grpcReg != nil && a.GRPC != nil {
		a.grpcReg(a.GRPC, infra)
	}
	if a.cfg.GRPC.Enabled && a.GRPC != nil && a.Registry != nil && a.cfg.AppName != "" {
		if err := a.GRPC.EnsureListen(); err != nil {
			a.log.Infof("discovery: grpc listen prepare: %v", err)
		} else {
			addr := discovery.ResolveAdvertiseAddr(a.cfg.AppName, a.GRPC.Addr(), a.cfg.Discovery)
			if addr != "" {
				if err := a.Registry.Register(context.Background(), a.cfg.AppName, addr); err != nil {
					a.log.Infof("discovery register %s -> %s: %v", a.cfg.AppName, addr, err)
				} else {
					a.log.Infof("discovery registered %s -> %s", a.cfg.AppName, addr)
				}
			}
		}
	}
	if a.httpReg != nil {
		if err := a.httpReg(a.engine, infra); err != nil {
			return fmt.Errorf("register http routes: %w", err)
		}
	}
	if a.Cron != nil {
		if err := a.Cron.Start(); err != nil {
			return fmt.Errorf("scheduler: %w", err)
		}
	}
	a.httpSrv = &http.Server{
		Addr:    a.cfg.HTTPAddr(),
		Handler: a.engine,
	}
	errCh := make(chan error, 2)
	if a.GRPC != nil {
		go func() {
			a.log.Infof("grpc listening on %s", a.GRPC.Addr())
			if err := a.GRPC.Serve(); err != nil {
				errCh <- fmt.Errorf("grpc: %w", err)
			}
		}()
	}
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
	a.closeAuditRecorder()
	if err := trace.Shutdown(ctx); err != nil {
		a.log.Infof("trace shutdown: %v", err)
	}
	if a.Cron != nil {
		_ = a.Cron.Stop()
	}
	if a.Registry != nil {
		if err := a.Registry.Deregister(ctx); err != nil {
			a.log.Infof("discovery deregister: %v", err)
		}
	}
	if a.GRPCRegistry != nil {
		if err := a.GRPCRegistry.Close(); err != nil {
			a.log.Infof("grpc registry close: %v", err)
		}
	}
	if a.GRPC != nil {
		_ = a.GRPC.Shutdown(ctx)
	}
	return a.httpSrv.Shutdown(ctx)
}

func (a *App) buildHTTPInfra() HTTPInfra {
	locker := cache.NewRedisLocker(cache.RedisClient(a.Cache), a.cfg.RedisKeyPrefix())
	a.auditRecorderRuntime = audit.BuildRecorder(a.cfg.Observability.Audit, a.auditRecorder)
	return HTTPInfra{
		DB:                 a.DB,
		DBDriver:           a.cfg.DB.Driver,
		TablePrefix:        a.cfg.DB.TablePrefix,
		Cache:              a.Cache,
		Locker:             locker,
		GuardedGet:         cache.NewGuardedGet(),
		RedisKeyPrefix:     a.cfg.RedisKeyPrefix(),
		EntityCacheEnabled: a.cfg.EntityCacheEnabled(),
		EntityCacheTTL:     a.cfg.EntityCacheTTL(),
		NullCacheTTL:       a.cfg.NullCacheTTL(),
		JWT:                a.JWT,
		MemberJWT:          a.MemberJWT,
		Casbin:             a.Casbin,
		APISign:            a.APISign,
		MQ:                 a.MQ,
		Outbox:             a.OutboxPub,
		MQLog:              a.cfg.Observability.MQ,
		EventBus:           a.EventBus,
		Cron:               a.Cron,
		GRPCClientForService: func(ctx context.Context, service string) (*grpcx.Client, error) {
			return a.NewGRPCClientForService(ctx, service)
		},
		RPC: a.GRPCRegistry,
		AuditRecorder: a.auditRecorderRuntime,
	}
}

func (a *App) closeAuditRecorder() {
	if a == nil || a.auditRecorderRuntime == nil {
		return
	}
	if c, ok := a.auditRecorderRuntime.(interface{ Close() }); ok {
		c.Close()
	}
}

// RunConsumer 初始化 MQ 消费者；可选启动 consumer_http 健康探针（不启动业务 HTTP/gRPC）。
func (a *App) RunConsumer() error {
	if a.consumerReg == nil {
		return errors.New("goeasy app: consumer register is nil")
	}
	gin.SetMode(gin.ReleaseMode)
	if err := a.InitInfra(); err != nil {
		return err
	}
	if !a.cfg.MQ.Enabled {
		return errors.New("goeasy app: mq.enabled must be true for consumer")
	}
	infra := a.buildHTTPInfra()
	a.consumerReg(a.MQ, infra)
	a.log.Infof("mq consumer running (lookupd=%s channel=%s)", a.cfg.MQ.LookupdAddress(), a.cfg.MQ.ConsumerChannel())

	errCh := make(chan error, 1)
	if a.cfg.ConsumerHTTP.Enabled {
		a.consumerHttpSrv = &http.Server{
			Addr:    a.cfg.ConsumerHTTPAddr(),
			Handler: a.engine,
		}
		go func() {
			a.log.Infof("consumer http listening on %s (health probe)", a.cfg.ConsumerHTTPAddr())
			if err := a.consumerHttpSrv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
				errCh <- fmt.Errorf("consumer http: %w", err)
			}
		}()
	} else {
		a.log.Infof("consumer_http.enabled is false; no health HTTP server started")
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	select {
	case err := <-errCh:
		return err
	case <-quit:
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	a.log.Infof("shutting down consumer...")
	if a.consumerHttpSrv != nil {
		_ = a.consumerHttpSrv.Shutdown(ctx)
	}
	if a.MQ != nil {
		_ = a.MQ.Close()
	}
	return nil
}
