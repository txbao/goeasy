# 07 DDD Lite 实践

本框架采用 **DDD Lite**：保留分层与边界，不强制事件溯源、复杂 CQRS 基础设施。

## 四层职责

| 层 | 目录 | 做什么 | 不做什么 |
|----|------|--------|----------|
| 领域 | `domain/` | 规则、实体行为、聚合不变量 | HTTP、SQL、第三方 SDK |
| 应用 | `app/` | 用例编排、事务边界、DTO | 直接写 SQL |
| 接口 | `interface/` | 协议适配、参数校验 | 业务规则 |
| 基础设施 | `infrastructure/` | 仓储实现、MQ、缓存客户端 | 领域规则 |

## 实体要有行为

反例：到处改 `status` 字段的贫血模型。

正例（health 示范）：

```go
func (s *ServiceHealth) MarkHealthy(message string) error {
    if message == "" {
        return ErrInvalidMessage
    }
    s.state = StateUp
    s.message = message
    return nil
}
```

对外只暴露方法或 `ToStatus()` 快照，不暴露可随意修改的字段。

## 聚合根

`aggregate.go` 负责维护聚合内一致性（例如根实体 + 子集合）。跨聚合修改通过应用服务或领域事件协调，不在一个 Repository 里隐式改两个聚合。

脚手架 `health` 模块用技术场景演示聚合，**不会**生成 User/Order 等业务模板。

## 领域服务

当规则涉及多个实体或不适合放在单个实体上时，使用 `domain/<m>/service.go`。

应用层调用顺序建议：

```text
Application → DomainService / Entity → Repository
```

读用例（Query）可以只读仓储返回 DTO，但写用例必须经过领域行为。

## 命令查询职责分离(CQRS) 目录约定

```text
app/<module>/
├── application.go      门面，对外统一入口
├── command_handler.go  写用例入口
├── query_handler.go    读用例入口
├── command/            每个写命令一个文件
├── query/              每个查询一个文件
├── dto.go
└── port.go             对外部系统端口（可选）
```

- **Command**：创建、更新、状态变更
- **Query**：列表、详情、报表

Handler（HTTP）只依赖 `Application` 或 Handler 结构体，不直接 new Repository。

## 依赖注入（bootstrap）

```go
func RegisterRoutes(engine *gin.Engine) {
    repo := healthinfra.NewMemoryRepository()
    domainSvc := domainhealth.NewDomainService()
    app := apphealth.NewApplication(repo, domainSvc)
    h := healthhttp.NewHandler(app)
    healthhttp.RegisterRoutes(engine, h)
}
```

新增业务模块时复制该模式，保持 **interface → app → domain ← infrastructure**。

## 新增业务模块推荐流程

1. `goesy add module <name>`
2. 实现 `domain/<name>` 实体、聚合、仓储接口
3. 实现 `infrastructure/persistence/<name>`
4. 编写 `app/<name>` 的 command/query
5. 添加 `interface/http/<name>` 与路由
6. 在 `bootstrap/wire.go` 装配并注册路由

## 脚手架边界

| 允许 | 禁止 |
|------|------|
| health 等技术示范 | CLI 默认生成 User、Order、Payment 等业务实体 |
| `add module` 由团队命名 | 在 framework 仓库写客户业务代码 |

## 延伸阅读

- [模板 v2 说明](../plans/goesy-cli-ddd-lite-template-v2.md)
- 架构规则：仓库 `.rulesync/rules/architecture/`

## 下一步

[08 项目模板](08-templates.md)
