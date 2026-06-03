# 常见问题

## CLI

### `goeasy.exe: command not found`

在 `goeasy-cli` 目录未加入 PATH。请使用：

```bat
go build -o goeasy.exe .
.\goeasy.exe version
```

或使用 `go install` 后的全局命令 `goeasy`。

### 未指定 `--module` 警告

重新创建时显式传入：

```bat
.\goeasy.exe new mysvc --module github.com/org/mysvc
```

已生成项目可手工修改 `go.mod` 第一行并全局替换 import 前缀。

### `go mod tidy` 拉取 goeasy 失败

- 检查 `GOPRIVATE` 与企业 Git 权限
- monorepo 内确认 `replace` 指向存在的 `goeasy` 目录

## 编译

### `go build` 找不到 goeasy 包

确认 `go.mod` 中 `require` 与 `replace` 正确；在业务项目根执行 `go mod tidy`。

### interface 层 import infrastructure 报错（架构守卫）

这是预期约束。将 `new Repository` 移到 `bootstrap/wire.go`，Handler 只接收 Application。

## 运行

### 端口占用

修改 `configs/config.yaml` 中 `server.port`。

### `/health` 404

确认 `bootstrap.RegisterRoutes` 已注册 health 路由，且服务已 `RegisterHTTP`。

### `/healthz` 无响应

在配置中启用 `observability.health.enabled`（参见 `config.example.yaml`）。

## 模板与升级

### 远端模板下载失败

CLI 会自动回退内嵌模板；日志中会有 fallback 提示。开发期建议 `--download=false`。

### 旧 demo 目录结构不一致

使用最新 CLI 重新 `goeasy new`，或对照 `goeasy-cli/demo6` 与 [04 项目结构](../04-project-structure.md) 手工迁移。

## 获取更多帮助

- [开发教程目录](../README.md)
- [goeasy 运行时规范](../../spec/goeasy-runtime-spec.md)
