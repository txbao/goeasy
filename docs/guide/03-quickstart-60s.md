# 03 60 秒上手

## 目标

创建名为 `greet` 的服务，启动后访问健康检查接口。

## 步骤

### 1. 进入脚手架目录并编译（若尚未编译）

```bat
cd goeasy-cli
go build -o goeasy.exe .
```

### 2. 创建项目

```bat
.\goeasy.exe new greet --module github.com/demo/greet --download=false
```

说明：

- `--module` 建议写真实 Go 模块路径，避免默认使用项目名 `greet`。
- `--download=false` 使用内嵌模板（默认）。

### 3. 进入项目并拉依赖

```bat
cd greet
go mod tidy
```

### 4. 启动服务

```bat
go run .\cmd\service
```

或先编译再运行：

```bat
go build -o greet.exe .\cmd\service
.\greet.exe
```

### 5. 验证接口

业务健康检查（DDD 示例模块）：

```bat
curl http://127.0.0.1:8080/health
```

预期 JSON 形如：

```json
{"code":0,"msg":"success","data":{"ok":true,"message":"greet is healthy"}}
```

框架探针（配置 `observability.health.enabled=true` 时）：

```bat
curl http://127.0.0.1:8080/healthz
```

## 生成后的入口长什么样

`cmd/service/main.go` 仅负责装配 goeasy：

```go
cfg := config.MustLoad("configs/config.yaml")
application := app.New(cfg)
application.RegisterHTTP(bootstrap.RegisterRoutes)
application.Run()
```

业务路由在 `internal/bootstrap/wire.go` 注册。

## 下一步

- 理解目录：[04 项目结构](04-project-structure.md)
- 命令大全：[06 goeasy-cli 命令](06-goeasy-cli-commands.md)
