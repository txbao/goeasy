# 10 架构对照

本章帮助从常见 Go 微服务实践迁移到 GoEasy，**不绑定任何第三方框架名称**。

## 三种常见形态

| 形态 | 典型目录 | 优点 | 风险 |
|------|----------|------|------|
| 三层 MVC | `controller/service/dao` | 上手快 | 业务与 DB 耦合，难以测试领域规则 |
| 按技术分包 | `handler/model/repo` | 文件好找 | 模块边界模糊，改一处牵全身 |
| DDD Lite（GoEasy） | `domain/app/interface/infrastructure` | 规则集中、可测、可替换基础设施 | 需要团队遵守依赖方向 |

GoEasy 选择第三种，并用 **goesy** 承担运行时横切能力。

## 启动方式对照

**传统手写 main：**

```text
读配置 → 创建 Gin → 散落注册路由 → ListenAndServe
```

**GoEasy：**

```text
config.MustLoad → app.New → RegisterHTTP(bootstrap) → Run（含优雅退出、观测钩子）
```

业务差异集中在 `bootstrap.RegisterRoutes`，main 保持极简。

## 代码生成对照

| 能力 | 传统手写 | GoEasy |
|------|----------|------|
| 新建服务 | 复制旧项目删代码 | `goesy new` |
| 新模块 | 复制粘贴目录 | `goesy add module` |
| API 契约 | 手写 proto | `goesy add proto` |
| 运行时 | 自选库拼装 | 统一 `goesy` P0–P4 |

CLI **不**生成业务实体；只提供 health 示范与 `add module` 骨架。

## 健康检查对照

| 类型 | 路径 | 归属 |
|------|------|------|
| 业务健康 | `GET /health` | `interface/http/health` + 领域状态 |
| 框架探针 | `GET /healthz` | `goesy/health`（配置开启） |

避免在业务 Handler 里重复实现 K8s 探针逻辑，优先用框架探针 + 业务 `/health` 组合。



## 回到教程首页

[开发教程目录](README.md)
