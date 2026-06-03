# goesy

GoEasy 企业级 Go 微服务运行时框架。开发教程见 [docs/guide](docs/guide/README.md)。

## 快速开始

```go
cfg := config.MustLoad("configs/config.yaml")
application := app.New(cfg)
application.RegisterHTTP(bootstrap.RegisterRoutes)
application.Run()
```

## 能力分层

| 层 | 包 |
|----|-----|
| P0 核心 | app, config, httpx, response, logger |
| P1 基础设施 | database, cache, mq, grpcx, discovery, storage, scheduler |
| P2 治理 | breaker, limiter, retry, loadbalance |
| P3 观测 | trace, metrics, health, audit |
| P4 企业组件 | errors, validator, pagination, idgen, contextx, jwt, casbin, crypto |

## 配置示例

```yaml
governance:
  limiter:
    enabled: true
    qps: 200
observability:
  metrics:
    enabled: true
    path: /metrics
  health:
    enabled: true
    path: /healthz
enterprise:
  jwt:
    enabled: true
    secret: change-me
```

## 开发指引

- [docs/guide/README.md](docs/guide/README.md)

## 废弃

`server`、`router`、`middleware` 根包为早期占位，请使用 `app` + `httpx`。
