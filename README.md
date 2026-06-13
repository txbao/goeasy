# goeasy

GoEasy 企业级 Go 微服务运行时框架。开发教程见 [goeasy-cli/docs](../goeasy-cli/docs/guide/README.md)（GitBook 同步源）。


## 安装

```bat
# 1. 安装 goeasy-cli（代码生成工具）
go install github.com/txbao/goeasy-cli@latest
goeasy-cli new demo --module github.com/demo/demo --download=false
cd demo
go mod tidy
go run ./cmd/service
```

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
| P4 企业组件 | errors, validator, pagination, idgen, contextx, jwt, casbin, crypto, apisign, eventbus |

### 操作日志（audit）

- **运维 JSON**：`observability.audit.enabled` → `app.Audit`（stdout）
- **业务持久化**：实现 `audit.Recorder`，`app.SetAuditRecorder(...)` 注入
- **上下文**：`httpx.InjectOperatorContext` + `contextx.OperatorFrom`
- 详见 [audit/README.md](audit/README.md)

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

- [GoEasy 开发文档](../goeasy-cli/docs/guide/README.md)（权威）
- [实体缓存](../goeasy-cli/docs/runtime/entity-cache.md)
- [HTTP 中间件](../goeasy-cli/docs/runtime/http-middleware.md)
- [gRPC 与服务发现](../goeasy-cli/docs/runtime/grpc-discovery.md)；业务 proto 见 [11 gRPC 项目集成](../goeasy-cli/docs/guide/11-grpc-internal.md)

## 废弃

`server`、`router`、`middleware` 根包为早期占位，请使用 `app` + `httpx`。
