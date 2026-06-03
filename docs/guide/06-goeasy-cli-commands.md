# 06 goeasy-cli 命令

CLI 二进制名为 **`goeasy`**（`go install` 后）。本地编译产物为 **`goeasy.exe`**，需在 `goeasy-cli` 目录用 `.\goeasy.exe` 调用。

## 命令总览

| 命令 | 说明 |
|------|------|
| `goeasy-cli version` | 版本信息 |
| `goeasy-cli new <name>` | 创建 DDD Lite 项目 |
| `goeasy-cli init <name>` | 同 `new` |
| `goeasy-cli add module <name>` | 生成完整业务模块骨架 |
| `goeasy-cli add crud <name>` | CRUD 占位文件 |
| `goeasy-cli add repository <name>` | 仓储接口 + infra 桩 |
| `goeasy-cli add proto <name>` | `api/proto` 定义 |
| `goeasy-cli add event <name>` | 领域事件 + 发布桩 |
| `goeasy-cli add aggregate <name>` | 聚合骨架（建议优先用 `add module`） |
| `goeasy-cli upgrade template` | 内嵌模板升级说明 |
| `goeasy-cli upgrade framework` | 查看 go.mod 中 goeasy 版本 |
| `goeasy-cli migrate up` | 应用 `migrations/*.up.sql` |
| `goeasy-cli migrate down --steps <n>` | 回滚最近 n 个迁移 |
| `goeasy-cli migrate status` | 查看迁移状态 |
| `goeasy-cli migrate version` | 查看当前迁移版本 |
| `goeasy-cli migrate goto <version>` | 迁移到指定版本 |
| `goeasy-cli migrate force <version>` | 强制设置版本，不执行 SQL |
| `goeasy-cli migrate create <name>` | 新建 up/down SQL 对 |

## new / init

```bat
goeasy-cli new mysvc --module github.com/org/mysvc --download=false
```

### 常用参数

| 参数 | 默认 | 说明 |
|------|------|------|
| `--module` | 项目名 | Go module 路径，**强烈建议显式指定** |
| `--template` | `default` | 见 [08 项目模板](08-templates.md) |
| `--version` | `v1.0.0` | 远端模板版本（配合 `--download`） |
| `--download` | `false` | `true` 时尝试拉远端，失败回退内嵌模板 |
| `--goeasy-module` | `github.com/txbao/goeasy` | 运行时 module（或 `GOEASY_MODULE`） |
| `--goeasy-replace` | 自动检测 | monorepo 内 replace 本地 goeasy |

未传 `--module` 时 CLI 会输出警告，仍可使用项目名作为 module。

## add（在已生成项目根目录）

```bat
cd mysvc
goeasy-cli add module order --dir .
```

| 参数 | 说明 |
|------|------|
| `--dir` | 项目根目录，默认当前目录 |
| `--force` | 覆盖已存在文件 |

`add module` 会生成完整 DDD 目录：`domain` + `app/command` + `app/query` + `interface/http` + `infrastructure/persistence`。

生成后需：

1. 在 `bootstrap/wire.go` 手工装配新模块（CLI 不自动改 wire，避免冲突）
2. `go mod tidy`

## migrate（在项目根目录）

```bat
goeasy-cli migrate up -f configs/config.yaml
goeasy-cli migrate status
goeasy-cli migrate version
goeasy-cli migrate goto 3
ngoeasy-cli migrate force 3
goeasy-cli migrate down --steps 1
goeasy-cli migrate create add_users_table
```

| 参数 | 默认 | 说明 |
|------|------|------|
| `--dir` | `.` | 项目根 |
| `-f` / `--config` | `configs/config.yaml` | 读取 `database.dsn` |
| `--migrations` | `migrations` | SQL 目录 |

需 `database.enabled: true` 且 `driver: postgres`。版本表为 `_sqlx_migrations`。

## upgrade

- `upgrade template`：模板随 CLI 版本发布，升级 CLI 即升级内嵌模板
- `upgrade framework`：提示业务项目 bump `goeasy` 依赖版本

## Windows 注意事项

| 现象 | 处理 |
|------|------|
| `goeasy-cli.exe: command not found` | 使用 `.\goeasy-cli.exe` 或把 `go\bin` 加入 PATH 后用 `goeasy-cli` |
| 在 `goeasy-cli` 目录未编译 | 先 `go build -o goeasy-cli.exe .` |

## 下一步

[07 DDD Lite 实践](07-ddd-lite-practices.md)
