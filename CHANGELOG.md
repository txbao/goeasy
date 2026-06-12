# Changelog

## Unreleased

### Added

- `response.FailBiz`：HTTP Status 与 6 位 `body.code` 分离（CSDS §9）
- `response.FailInternal`：未映射内部错误兜底（`500001`），prod 响应脱敏
- `response.Configure` / `SetLogger` / `SetEnv`：httpx 启动时注入 logger 与环境
- `errors.BizError` 与全局保留业务码（`100001`/`100002`/`200002`/`210001`/`500001`/`503001`）
- `httpCode >= 500` 或 `bizCode >= 500000` 时自动 `slog.Error`（含 `request_id`/`trace_id`/`user_id`）
- `config.observability.http` 可选配置（`log_server_errors`、`expose_error_detail`）

### Changed

- `httpx` 鉴权中间件（JWT/Casbin/APISign）迁移 `FailBiz` / `FailInternal`
- 访问日志 `status >= 500` 升为 `Warn`

### Deprecated

- `response.Fail`：保留兼容，`body.code` 仍等于 HTTP 状态码；新业务请用 `FailBiz`

### Upgrade（业务项目）

1. `go get github.com/txbao/goeasy@<版本>`
2. 删除临时 `internal/interface/http/respond/`（若有）
3. Handler：`mapErr` → `FailBiz`；未映射错误 → `FailInternal`
