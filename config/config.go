package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config 服务运行时配置（P0–P4）。
type Config struct {
	AppName       string        `yaml:"app_name"`
	Env           string        `yaml:"env"`
	HTTP          HTTP          `yaml:"http"`
	ConsumerHTTP  ConsumerHTTP  `yaml:"consumer_http"`
	DB            DB            `yaml:"database"`
	Redis         Redis         `yaml:"redis"`
	Cache         CacheCfg      `yaml:"cache"`
	MQ            MQ            `yaml:"mq"`
	GRPC          GRPC          `yaml:"grpc"`
	Discovery     Discovery     `yaml:"discovery"`
	Governance    Governance    `yaml:"governance"`
	Observability Observability `yaml:"observability"`
	Scheduler     SchedulerCfg  `yaml:"scheduler"`
	Enterprise    Enterprise    `yaml:"enterprise"`
}

type HTTP struct {
	Host string  `yaml:"host"`
	Port int     `yaml:"port"`
	CORS CORSCfg `yaml:"cors"`
}

// ConsumerHTTP cmd/consumer 独立进程的健康探针 HTTP（与业务 http 端口分离）。
type ConsumerHTTP struct {
	Enabled bool   `yaml:"enabled"`
	Host    string `yaml:"host"`
	Port    int    `yaml:"port"`
}

// CORSCfg 跨域（由 httpx.CORSMiddleware 消费）。
type CORSCfg struct {
	Enabled          bool     `yaml:"enabled"`
	AllowOrigins     []string `yaml:"allow_origins"`
	AllowMethods     []string `yaml:"allow_methods"`
	AllowHeaders     []string `yaml:"allow_headers"`
	AllowCredentials bool     `yaml:"allow_credentials"`
}

type DB struct {
	Enabled     bool   `yaml:"enabled"`
	ORM         string `yaml:"orm"` // sqlx（执行）+ goqu（SQL 构建）| gorm | ent
	DSN         string `yaml:"dsn"`
	Driver      string `yaml:"driver"`
	TablePrefix string `yaml:"table_prefix"`
	MaxOpen     int    `yaml:"max_open"`
	MaxIdle     int    `yaml:"max_idle"`
}

type Redis struct {
	Enabled    bool   `yaml:"enabled"`
	Addr       string `yaml:"addr"`
	Password   string `yaml:"password"`
	DB         int    `yaml:"db"`
	KeyPrefix  string `yaml:"key_prefix"`  // 项目级 Redis key 前缀，空则回退 app_name
	DefaultTTL string `yaml:"default_ttl"` // 如 168h，供实体缓存默认 TTL
}

// CacheCfg 实体缓存开关（需 redis.enabled；仓储层按此决定是否读写缓存）。
type CacheCfg struct {
	Enabled   bool       `yaml:"enabled"`
	EntityTTL string     `yaml:"entity_ttl"` // 如 168h，空则使用 redis.default_ttl
	L1        L1CacheCfg `yaml:"l1"`         // 进程内 L1，需 redis.enabled 作为 L2
	NullTTL   string     `yaml:"null_ttl"`   // 空值缓存 TTL，防穿透
	TTLJitter float64    `yaml:"ttl_jitter"` // 0~1，TTL 随机抖动比例，防雪崩
}

// L1CacheCfg 本地内存缓存（L1）。
type L1CacheCfg struct {
	Enabled    bool   `yaml:"enabled"`
	MaxEntries int    `yaml:"max_entries"`
	TTL        string `yaml:"ttl"` // 如 1m
}

// SchedulerCfg 定时任务。
type SchedulerCfg struct {
	Enabled  bool   `yaml:"enabled"`
	Timezone string `yaml:"timezone"` // 如 Asia/Shanghai，空为 Local
}

type MQ struct {
	Enabled     bool       `yaml:"enabled"`
	Type        string     `yaml:"type"`         // nsq
	Addr        string     `yaml:"addr"`         // 兼容旧字段，Producer 回退地址
	NSQDAddr    string     `yaml:"nsqd_addr"`    // Producer 直连 nsqd
	LookupdAddr string     `yaml:"lookupd_addr"` // Consumer lookupd
	Channel     string     `yaml:"channel"`      // 消费 channel，默认 default
	Outbox      OutboxCfg  `yaml:"outbox"`       // 事务发件箱，默认关闭
}

// OutboxCfg MQ 事务一致性（Outbox 模式）。
type OutboxCfg struct {
	Enabled      bool   `yaml:"enabled"`
	Table        string `yaml:"table"`
	PollInterval string `yaml:"poll_interval"` // 如 5s
	BatchSize    int    `yaml:"batch_size"`
	MaxRetries   int    `yaml:"max_retries"`
}

// NSQDAddress 返回 Producer 使用的 nsqd 地址。
func (m MQ) NSQDAddress() string {
	if m.NSQDAddr != "" {
		return m.NSQDAddr
	}
	return m.Addr
}

// LookupdAddress 返回 Consumer 使用的 lookupd 地址。
func (m MQ) LookupdAddress() string {
	if m.LookupdAddr != "" {
		return m.LookupdAddr
	}
	return m.Addr
}

// ConsumerChannel 返回 NSQ 消费 channel。
func (m MQ) ConsumerChannel() string {
	if m.Channel != "" {
		return m.Channel
	}
	return "default"
}

type GRPC struct {
	Enabled    bool   `yaml:"enabled"`
	Addr       string `yaml:"addr"`
	TimeoutSec int    `yaml:"timeout_sec"`
	MaxRetries int    `yaml:"max_retries"`
}

// Discovery 服务发现（direct 静态地址 | etcd）。
type Discovery struct {
	Mode     string            `yaml:"mode"`     // direct | etcd
	Etcd     EtcdDiscoveryCfg  `yaml:"etcd"`
	Services map[string]string `yaml:"services"` // direct：逻辑服务名 -> host:port
}

// EtcdDiscoveryCfg etcd 服务注册（optional）。
type EtcdDiscoveryCfg struct {
	Enabled        bool     `yaml:"enabled"`
	Endpoints      []string `yaml:"endpoints"`
	DialTimeoutSec int      `yaml:"dial_timeout_sec"`
	LeaseTTLSec    int      `yaml:"lease_ttl_sec"`   // 租约 TTL，默认 30；KeepAlive 保活至进程退出
	AdvertiseAddr  string   `yaml:"advertise_addr"` // 注册到 etcd 的可达地址，空则回退 services[app_name] 或监听地址
	Prefix         string   `yaml:"prefix"`          // 默认 /goeasy/services
}

// Governance P2 微服务治理配置。
type Governance struct {
	Breaker BreakerCfg `yaml:"breaker"`
	Limiter LimiterCfg `yaml:"limiter"`
	Retry   RetryCfg   `yaml:"retry"`
}

type BreakerCfg struct {
	Enabled     bool   `yaml:"enabled"`
	MaxRequests uint32 `yaml:"max_requests"`
	IntervalSec int    `yaml:"interval_sec"`
	TimeoutSec  int    `yaml:"timeout_sec"`
}

type LimiterCfg struct {
	Enabled    bool     `yaml:"enabled"`
	Mode       string   `yaml:"mode"`       // local | redis | both
	QPS        float64  `yaml:"qps"`
	Burst      int      `yaml:"burst"`
	Dimensions []string `yaml:"dimensions"` // global, ip, user
}

type RetryCfg struct {
	Enabled    bool `yaml:"enabled"`
	MaxRetries int  `yaml:"max_retries"`
	BackoffMs  int  `yaml:"backoff_ms"`
}

// Observability P3 观测配置。
type Observability struct {
	Logger  LoggerCfg  `yaml:"logger"`
	SQL     SQLCfg     `yaml:"sql"`
	Trace   TraceCfg   `yaml:"trace"`
	Metrics MetricsCfg `yaml:"metrics"`
	Audit   AuditCfg   `yaml:"audit"`
	Health  HealthCfg  `yaml:"health"`
	MQ      MQLogCfg   `yaml:"mq"`
}

// LoggerCfg 结构化日志。
type LoggerCfg struct {
	Level  string `yaml:"level"`  // debug | info | warn | error
	Format string `yaml:"format"` // json | text
	Output string `yaml:"output"` // stdout | stderr
}

// SQLCfg SQL 日志与慢查询。
type SQLCfg struct {
	Enabled bool `yaml:"enabled"`
	SlowMs  int  `yaml:"slow_ms"`
}

// MQLogCfg MQ 发布/消费日志（payload 预览默认关闭）。
type MQLogCfg struct {
	LogPayload         bool `yaml:"log_payload"`
	LogPayloadMaxBytes int  `yaml:"log_payload_max_bytes"`
}

type TraceCfg struct {
	Enabled     bool    `yaml:"enabled"`
	Service     string  `yaml:"service"`
	Exporter    string  `yaml:"exporter"` // noop | otlp
	Protocol    string  `yaml:"protocol"` // grpc | http
	Endpoint    string  `yaml:"endpoint"`
	Insecure    bool    `yaml:"insecure"`
	SampleRatio float64 `yaml:"sample_ratio"` // 0~1
}

type MetricsCfg struct {
	Enabled bool   `yaml:"enabled"`
	Path    string `yaml:"path"`
}

type AuditCfg struct {
	Enabled bool `yaml:"enabled"`
}

type HealthCfg struct {
	Enabled bool   `yaml:"enabled"`
	Path    string `yaml:"path"`
}

// Enterprise P4 企业公共组件配置。
type Enterprise struct {
	JWT       JWTCfg     `yaml:"jwt"`
	MemberJWT JWTCfg     `yaml:"member_jwt"` // H5 会员 JWT，与后台 jwt 独立，按需 enabled
	Casbin    CasbinCfg  `yaml:"casbin"`
	APISign   APISignCfg `yaml:"api_sign"`
	IDGen     IDGenCfg   `yaml:"idgen"`
}

// APISignCfg 开放平台 RSA2 签名验签（见 api-sign.md）。
type APISignCfg struct {
	Enabled         bool                      `yaml:"enabled"`
	TimestampSkewMs int64                     `yaml:"timestamp_skew_ms"`
	Apps            map[string]APISignAppCfg  `yaml:"apps"`
}

// APISignAppCfg 接入方公钥。
type APISignAppCfg struct {
	PublicKeyPEM string `yaml:"public_key_pem"`
}

type JWTCfg struct {
	Enabled   bool   `yaml:"enabled"`
	Secret    string `yaml:"secret"`
	Issuer    string `yaml:"issuer"`
	ExpireMin int    `yaml:"expire_min"`
}

type CasbinCfg struct {
	Enabled bool   `yaml:"enabled"`
	Model   string `yaml:"model"`
	Policy  string `yaml:"policy"`
}

type IDGenCfg struct {
	Mode string `yaml:"mode"` // uuid | snowflake
}

func Load(path string) (*Config, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := yaml.Unmarshal(b, &cfg); err != nil {
		return nil, err
	}
	applyDefaults(&cfg)
	return &cfg, nil
}

func MustLoad(path string) *Config {
	cfg, err := Load(path)
	if err != nil {
		panic(fmt.Sprintf("goeasy config: %v", err))
	}
	return cfg
}

func applyDefaults(cfg *Config) {
	if cfg.HTTP.Host == "" {
		cfg.HTTP.Host = "0.0.0.0"
	}
	if cfg.HTTP.Port == 0 {
		cfg.HTTP.Port = 8080
	}
	if cfg.ConsumerHTTP.Host == "" {
		cfg.ConsumerHTTP.Host = "0.0.0.0"
	}
	if cfg.ConsumerHTTP.Port == 0 {
		cfg.ConsumerHTTP.Port = 18080
	}
	if cfg.DB.Driver == "" {
		cfg.DB.Driver = "postgres"
	}
	if cfg.DB.ORM == "" {
		cfg.DB.ORM = "sqlx"
	}
	if cfg.Governance.Breaker.MaxRequests == 0 {
		cfg.Governance.Breaker.MaxRequests = 3
	}
	if cfg.Governance.Breaker.IntervalSec == 0 {
		cfg.Governance.Breaker.IntervalSec = 60
	}
	if cfg.Governance.Breaker.TimeoutSec == 0 {
		cfg.Governance.Breaker.TimeoutSec = 30
	}
	if cfg.Governance.Limiter.QPS == 0 {
		cfg.Governance.Limiter.QPS = 100
	}
	if cfg.Governance.Limiter.Burst == 0 {
		cfg.Governance.Limiter.Burst = 200
	}
	if cfg.Governance.Limiter.Mode == "" {
		cfg.Governance.Limiter.Mode = "local"
	}
	if len(cfg.Governance.Limiter.Dimensions) == 0 {
		cfg.Governance.Limiter.Dimensions = []string{"global"}
	}
	if cfg.Observability.Logger.Level == "" {
		if cfg.Env == "dev" {
			cfg.Observability.Logger.Level = "debug"
		} else {
			cfg.Observability.Logger.Level = "info"
		}
	}
	if cfg.Observability.Logger.Format == "" {
		cfg.Observability.Logger.Format = "json"
	}
	if cfg.Observability.Logger.Output == "" {
		cfg.Observability.Logger.Output = "stdout"
	}
	if cfg.Observability.SQL.SlowMs == 0 {
		cfg.Observability.SQL.SlowMs = 200
	}
	if cfg.Cache.L1.MaxEntries == 0 {
		cfg.Cache.L1.MaxEntries = 10000
	}
	if cfg.Cache.L1.TTL == "" {
		cfg.Cache.L1.TTL = "1m"
	}
	if cfg.Cache.NullTTL == "" {
		cfg.Cache.NullTTL = "5m"
	}
	if cfg.Enterprise.APISign.TimestampSkewMs == 0 {
		cfg.Enterprise.APISign.TimestampSkewMs = 300000
	}
	if cfg.Enterprise.APISign.Apps == nil {
		cfg.Enterprise.APISign.Apps = make(map[string]APISignAppCfg)
	}
	if cfg.Governance.Retry.MaxRetries == 0 {
		cfg.Governance.Retry.MaxRetries = 3
	}
	if cfg.Governance.Retry.BackoffMs == 0 {
		cfg.Governance.Retry.BackoffMs = 100
	}
	if cfg.Observability.Metrics.Path == "" {
		cfg.Observability.Metrics.Path = "/metrics"
	}
	if cfg.Observability.Health.Path == "" {
		cfg.Observability.Health.Path = "/healthz"
	}
	if cfg.Observability.MQ.LogPayloadMaxBytes == 0 {
		cfg.Observability.MQ.LogPayloadMaxBytes = 512
	}
	if cfg.Enterprise.IDGen.Mode == "" {
		cfg.Enterprise.IDGen.Mode = "uuid"
	}
	if cfg.Enterprise.JWT.ExpireMin == 0 {
		cfg.Enterprise.JWT.ExpireMin = 60
	}
	if cfg.GRPC.TimeoutSec == 0 {
		cfg.GRPC.TimeoutSec = 5
	}
	if cfg.GRPC.MaxRetries == 0 {
		cfg.GRPC.MaxRetries = 2
	}
	if cfg.Redis.DefaultTTL == "" {
		cfg.Redis.DefaultTTL = "168h"
	}
	if cfg.Enterprise.MemberJWT.ExpireMin == 0 {
		cfg.Enterprise.MemberJWT.ExpireMin = 10080 // 7 天
	}
	if cfg.Discovery.Mode == "" {
		cfg.Discovery.Mode = "direct"
	}
	if cfg.Discovery.Etcd.Prefix == "" {
		cfg.Discovery.Etcd.Prefix = "/goeasy/services"
	}
	if cfg.Discovery.Etcd.DialTimeoutSec == 0 {
		cfg.Discovery.Etcd.DialTimeoutSec = 5
	}
	if cfg.Discovery.Etcd.LeaseTTLSec == 0 {
		cfg.Discovery.Etcd.LeaseTTLSec = 30
	}
	if cfg.Discovery.Services == nil {
		cfg.Discovery.Services = make(map[string]string)
	}
	if cfg.Observability.Trace.Exporter == "" {
		cfg.Observability.Trace.Exporter = "noop"
	}
	if cfg.Observability.Trace.Protocol == "" {
		cfg.Observability.Trace.Protocol = "grpc"
	}
	if cfg.Observability.Trace.SampleRatio <= 0 {
		cfg.Observability.Trace.SampleRatio = 1
	}
	if cfg.Scheduler.Timezone == "" {
		cfg.Scheduler.Timezone = "Asia/Shanghai"
	}
	if cfg.Cache.TTLJitter <= 0 {
		cfg.Cache.TTLJitter = 0.1
	}
	if cfg.MQ.Outbox.Table == "" {
		cfg.MQ.Outbox.Table = "goeasy_outbox"
	}
	if cfg.MQ.Outbox.PollInterval == "" {
		cfg.MQ.Outbox.PollInterval = "5s"
	}
	if cfg.MQ.Outbox.BatchSize == 0 {
		cfg.MQ.Outbox.BatchSize = 50
	}
	if cfg.MQ.Outbox.MaxRetries == 0 {
		cfg.MQ.Outbox.MaxRetries = 5
	}
}

// TraceService 返回 trace 服务名（空则用 app_name）。
func (c *Config) TraceService() string {
	if c == nil {
		return "goeasy"
	}
	if s := c.Observability.Trace.Service; s != "" {
		return s
	}
	if c.AppName != "" {
		return c.AppName
	}
	return "goeasy"
}

// OutboxTable 返回带 table_prefix 的 outbox 表名。
func (c *Config) OutboxTable() string {
	if c == nil {
		return "goeasy_outbox"
	}
	t := c.MQ.Outbox.Table
	if t == "" {
		t = "goeasy_outbox"
	}
	if p := c.DB.TablePrefix; p != "" {
		return p + t
	}
	return t
}

// OutboxPollInterval 解析 outbox relay 轮询间隔。
func (c *Config) OutboxPollInterval() time.Duration {
	if c == nil {
		return 5 * time.Second
	}
	raw := c.MQ.Outbox.PollInterval
	if raw == "" {
		raw = "5s"
	}
	d, err := time.ParseDuration(raw)
	if err != nil {
		return 5 * time.Second
	}
	return d
}

// RedisKeyPrefix 返回 Redis 实体 key 的项目前缀（key_prefix 为空时用 app_name）。
func (c *Config) RedisKeyPrefix() string {
	if c == nil {
		return ""
	}
	if p := c.Redis.KeyPrefix; p != "" {
		return p
	}
	return c.AppName
}

// EntityCacheTTL 解析实体缓存 TTL（cache.entity_ttl → redis.default_ttl → 168h）。
func (c *Config) EntityCacheTTL() time.Duration {
	if c == nil {
		return 168 * time.Hour
	}
	raw := c.Cache.EntityTTL
	if raw == "" {
		raw = c.Redis.DefaultTTL
	}
	if raw == "" {
		raw = "168h"
	}
	d, err := time.ParseDuration(raw)
	if err != nil {
		return 168 * time.Hour
	}
	return d
}

// EntityCacheEnabled 是否启用仓储实体缓存（redis 与 cache 均需 enabled）。
func (c *Config) EntityCacheEnabled() bool {
	if c == nil {
		return false
	}
	return c.Redis.Enabled && c.Cache.Enabled
}

// L1CacheTTL 解析 L1 缓存 TTL。
func (c *Config) L1CacheTTL() time.Duration {
	if c == nil {
		return time.Minute
	}
	raw := c.Cache.L1.TTL
	if raw == "" {
		raw = "1m"
	}
	d, err := time.ParseDuration(raw)
	if err != nil {
		return time.Minute
	}
	return d
}

// NullCacheTTL 空值缓存 TTL（防穿透）。
func (c *Config) NullCacheTTL() time.Duration {
	if c == nil {
		return 5 * time.Minute
	}
	raw := c.Cache.NullTTL
	if raw == "" {
		raw = "5m"
	}
	d, err := time.ParseDuration(raw)
	if err != nil {
		return 5 * time.Minute
	}
	return d
}

func (c *Config) HTTPAddr() string {
	return fmt.Sprintf("%s:%d", c.HTTP.Host, c.HTTP.Port)
}

// ConsumerHTTPAddr 返回 consumer 健康探针监听地址。
func (c *Config) ConsumerHTTPAddr() string {
	if c == nil {
		return "0.0.0.0:18080"
	}
	host := c.ConsumerHTTP.Host
	if host == "" {
		host = "0.0.0.0"
	}
	port := c.ConsumerHTTP.Port
	if port == 0 {
		port = 18080
	}
	return fmt.Sprintf("%s:%d", host, port)
}
