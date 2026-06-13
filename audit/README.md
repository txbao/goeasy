# audit — 运维审计与业务操作日志

GoEasy 提供两条并存的审计能力，**不承载业务语义、不固化业务表**。

## 1. 运维 JSON 审计（`audit.Logger`）

配置 `observability.audit.enabled: true` 时，向 stdout 输出结构化 JSON：

```go
app.Audit.Record(ctx, audit.Record{
    Operator: "admin-1",
    Action:   "update",
    Resource: "config/rate_limit",
    IP:       "10.0.0.1",
    OldValue: map[string]any{"qps": 100},
    NewValue: map[string]any{"qps": 200},
})
```

`enabled: false` 时零开销（`inner == nil`）。

## 2. 业务操作日志 Port（`audit.Recorder`）

业务服务实现 `Recorder` 接口，写入自有表（如 `base_operation_logs`）：

```go
type Recorder interface {
    Record(ctx context.Context, op contextx.OperatorContext, entry Entry) error
}
```

注入方式（在 `Run` 前）：

```go
application.SetAuditRecorder(operationlogs.NewPGRecorder(dbx))
```

`HTTPInfra.AuditRecorder` 在 bootstrap 中可用。默认 `NopRecorder`；`async_enabled: true` 时自动包装 `AsyncRecorder`。

## 3. 操作人上下文

JWT 路由组挂载 `httpx.InjectOperatorContext()` 后：

```go
op := contextx.OperatorFrom(ctx)
recorder.Record(ctx, op, audit.Entry{...})
```

中间件顺序：`requestID → trace → RequireJWT → InjectOperatorContext → 业务路由`。

## 4. 脱敏与变更摘要

```go
opts := audit.RedactOptionsFromCfg(cfg.Observability.Audit)
before, after, _ := audit.BuildChangeSummary(old, new, []string{"customerName", "status"}, audit.SummaryOptions{Redact: opts})
```

内置敏感 key：`password`、`token`、`secret` 等；手机号保留前 3 后 4。

## 5. 明确不做

- 不自动将每个 HTTP 请求写入业务操作日志
- 不硬编码 `module_code` / `action_type` 业务枚举
- 不实现 List/Get/Export API
