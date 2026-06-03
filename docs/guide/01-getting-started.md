# 01 入门

## 概述

GoEasy 是企业级 Go 微服务技术栈，由两部分组成：

| 组件 | 职责 | 是否参与线上运行 |
|------|------|------------------|
| **goesy** | 运行时：配置、HTTP、日志、治理、观测、企业公共能力 | 是 |
| **goesy-cli** | 脚手架：创建项目、生成模块与契约 | 否（仅开发期） |
| **业务服务** | 领域模型、业务规则、业务流程 | 是 |

业务代码只写在业务服务仓库中；框架不提供 User、Order 等业务实体，仅提供可复用的技术能力。

## 核心思路

1. 用 **goesy-cli** 生成符合 **DDD Lite** 目录规范的项目骨架。
2. 用 **goesy/app** 统一启动 HTTP 服务、加载配置、挂载中间件。
3. 在 **bootstrap** 中完成依赖注入；**interface** 层只调 **application**，不直接 new 仓储实现。

典型启动代码：

```go
cfg := config.MustLoad("configs/config.yaml")
application := app.New(cfg)
application.RegisterHTTP(bootstrap.RegisterRoutes)
application.Run()
```

## 与「传统三层」的差异

传统 `controller → service → dao` 容易把业务与数据库耦在一起。DDD Lite 强调：

- **domain**：业务规则与实体行为
- **app**：用例编排（CQRS 读写分离）
- **interface**：HTTP/gRPC 适配
- **infrastructure**：仓储与外部系统实现

详见 [07 DDD Lite 实践](07-ddd-lite-practices.md)。

## 下一步

- 安装工具：[02 安装](02-installation.md)
- 跑通第一个服务：[03 60 秒上手](03-quickstart-60s.md)
