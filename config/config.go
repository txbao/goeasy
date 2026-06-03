package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config 服务运行时配置（P0–P4）。
type Config struct {
	AppName       string        `yaml:"app_name"`
	Env           string        `yaml:"env"`
	HTTP          HTTP          `yaml:"http"`
	DB            DB            `yaml:"database"`
	Redis         Redis         `yaml:"redis"`
	MQ            MQ            `yaml:"mq"`
	GRPC          GRPC          `yaml:"grpc"`
	Governance    Governance    `yaml:"governance"`
	Observability Observability `yaml:"observability"`
	Enterprise    Enterprise    `yaml:"enterprise"`
}

type HTTP struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

type DB struct {
	Enabled bool   `yaml:"enabled"`
	DSN     string `yaml:"dsn"`
	Driver  string `yaml:"driver"`
	MaxOpen int    `yaml:"max_open"`
	MaxIdle int    `yaml:"max_idle"`
}

type Redis struct {
	Enabled  bool   `yaml:"enabled"`
	Addr     string `yaml:"addr"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
}

type MQ struct {
	Enabled bool   `yaml:"enabled"`
	Type    string `yaml:"type"`
	Addr    string `yaml:"addr"`
}

type GRPC struct {
	Enabled    bool   `yaml:"enabled"`
	Addr       string `yaml:"addr"`
	TimeoutSec int    `yaml:"timeout_sec"`
	MaxRetries int    `yaml:"max_retries"`
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
	Enabled bool    `yaml:"enabled"`
	QPS     float64 `yaml:"qps"`
	Burst   int     `yaml:"burst"`
}

type RetryCfg struct {
	Enabled    bool `yaml:"enabled"`
	MaxRetries int  `yaml:"max_retries"`
	BackoffMs  int  `yaml:"backoff_ms"`
}

// Observability P3 观测配置。
type Observability struct {
	Trace   TraceCfg   `yaml:"trace"`
	Metrics MetricsCfg `yaml:"metrics"`
	Audit   AuditCfg   `yaml:"audit"`
	Health  HealthCfg  `yaml:"health"`
}

type TraceCfg struct {
	Enabled bool   `yaml:"enabled"`
	Service string `yaml:"service"`
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
	JWT    JWTCfg    `yaml:"jwt"`
	Casbin CasbinCfg `yaml:"casbin"`
	IDGen  IDGenCfg  `yaml:"idgen"`
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
		panic(fmt.Sprintf("goesy config: %v", err))
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
	if cfg.DB.Driver == "" {
		cfg.DB.Driver = "postgres"
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
}

func (c *Config) HTTPAddr() string {
	return fmt.Sprintf("%s:%d", c.HTTP.Host, c.HTTP.Port)
}
