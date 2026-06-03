# 名词表

| 术语 | 含义 |
|------|------|
| GoEasy | 企业级 Go 微服务技术栈（运行时 + 脚手架 + 规范） |
| goesy | 运行时 Go 模块，`github.com/txbao/goeasy` |
| goesy-cli | 脚手架 CLI，安装后命令名为 `goesy` |
| DDD Lite | 简化领域驱动设计：分层 + 实体行为 + CQRS 目录，不强制事件溯源 |
| Application | 应用层门面，编排领域对象完成用例 |
| Aggregate | 聚合根，维护一致性边界的入口实体 |
| Domain Service | 跨实体或不属于单实体的领域规则 |
| Repository | 仓储接口，定义在 domain，实现在 infrastructure |
| Command / Query | 写用例 / 读用例，分目录存放 |
| bootstrap | 依赖注入与 HTTP 路由注册，唯一允许大量 `new` 的位置 |
| interface 层 | 对外协议适配（HTTP、gRPC），不是 Go 的 `interface` 关键字 |
| health 模块 | 脚手架内置技术示范，不代表真实业务域 |
| P0–P4 | goesy 运行时能力分期（核心 → 基础设施 → 治理 → 观测 → 企业公共） |
| replace | go.mod 本地路径替换，用于 monorepo 联调 |
| embed 模板 | CLI 内嵌的脚手架文件，不依赖网络 |
