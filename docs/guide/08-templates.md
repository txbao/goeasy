# 08 项目模板

`goeasy new` 通过 `--template` 选择项目变体，均基于同一 DDD Lite 骨架，差异主要在配置默认值与 README 说明。

## 可用模板

| 名称 | 适用场景 |
|------|----------|
| `default` | 标准微服务（默认） |
| `monolith` | 单体部署，简化部分运维配置说明 |
| `auth` | 认证相关服务（预留 JWT/Casbin 配置项说明） |
| `system` | 系统/平台类服务 |
| `payment` | 支付类服务（强调审计与幂等配置占位） |

## 使用示例

```bat
.\goeasy.exe new auth-svc --template auth --module github.com/demo/auth --download=false
.\goeasy.exe new pay-svc --template payment --module github.com/demo/pay
.\goeasy.exe new app --template monolith --module github.com/demo/app
```

## 内嵌 vs 远端

| 模式 | 参数 | 行为 |
|------|------|------|
| 内嵌（推荐） | `--download=false` | 使用 CLI 内置 embed 模板，离线可用 |
| 远端 | `--download=true` | 尝试按 `--version` 下载，失败自动回退内嵌 |

日常开发与 CI 建议使用内嵌模板，保证与当前 CLI 版本一致。

## 模板内容（各变体共有）

- `cmd/service` + `internal/bootstrap/wire.go`
- `domain/health` 技术示范（实体行为、聚合、领域服务、CQRS 目录）
- `configs/config.yaml` + `config.example.yaml`
- `deploy/docker/Dockerfile`、目录占位

变体**不会**自动带入真实业务表结构或第三方密钥。

## 升级模板

```bat
go install github.com/txbao/goeasy-cli@latest
goeasy upgrade template
```

已有业务项目不会自动重写；需在业务仓库手工对比模板变更或重新生成模块文件。

## 下一步

[09 Monorepo 与 Module](09-monorepo-and-modules.md)
