# 09 Monorepo 与 Module

## Go Module 路径

创建项目时务必指定真实模块路径：

```bat
.\goesy.exe new ordersvc --module github.com/yourorg/ordersvc
```

生成后 `go.mod` 首行为：

```text
module github.com/yourorg/ordersvc
```

所有 `internal/...` 的 import 均基于该前缀。

## framework monorepo 开发

仓库布局示例：

```text
framework/
├── goesy/           github.com/txbao/goeasy
├── goesy-cli/       github.com/txbao/goeasy-cli
└── goesy-cli/demo/ 本地验收工程
```

在 `framework` 根目录执行 `goesy new` 时，CLI 会检测 monorepo 并写入：

```text
replace github.com/txbao/goeasy => ../goesy
```

（实际相对路径以生成时检测结果为准。）

也可手动指定：

```bat
.\goesy.exe new demo --module github.com/demo/demo --goesy-replace ..\goesy
```

## 业务独立仓库

业务代码单独建仓时：

1. 去掉或不要提交 `replace`（仅本地开发临时使用）
2. `require github.com/txbao/goeasy v0.x.x` 指向已发布版本
3. 配置 GOPRIVATE / 企业 Git 凭据拉取私有模块

```bat
go env -w GOPRIVATE=github.com/*
go mod tidy
```

## 多服务仓库（multi-module）

若一个 Git 仓库存多个服务，推荐每个服务独立 `go.mod`：

```text
repo/
├── service-a/go.mod
└── service-b/go.mod
```

每个服务目录内单独执行 `goesy new` 或复制已生成骨架，避免单一 module 路径混乱。

## 版本对齐

| 组件 | 建议 |
|------|------|
| goesy-cli | 团队统一 `go install` 版本 |
| goesy 运行时 | `go.mod` 与生产环境一致，用 `goesy upgrade framework` 核对 |

## 下一步

[10 架构对照](10-architecture-overview.md)
