# courier

[![GoDoc Widget](https://godoc.org/github.com/octohelm/courier?status.svg)](https://pkg.go.dev/github.com/octohelm/courier)

`courier` 是一个 Go 服务框架基础库，围绕类型化 operator、HTTP 承载、OpenAPI 描述生成、校验、错误表达和 client 调用提供可复用能力。

仓库同时包含核心库实现、开发期生成器、文档和可运行示例，适合作为服务框架底座或上层业务框架的依赖。

## 职责与边界

- root README 负责仓库概述、目录职责和继续阅读入口。
- 协作约束看 [AGENTS.md](AGENTS.md)。
- 执行入口看 [justfile](justfile) 和 [`internal/example/cmd/example/justfile`](internal/example/cmd/example/justfile)。

## 仓库导览

- [`pkg/`](pkg/)：核心公共库，包含 `courier` 抽象、HTTP 承载、OpenAPI、校验、状态错误与表达式等能力。
- [`devpkg/`](devpkg/)：开发期生成器与辅助包，用于支撑 operator、client、injectable 等代码生成。
- [`tool/internal/cmd/gen`](tool/internal/cmd/gen/)：仓库内统一使用的 `go tool gen` 入口。
- [`internal/example/`](internal/example/)：仓库内示例实现，展示契约层、endpoint 层、routes 组装层和注入式 service 的推荐分层。
- [`.agents/skills/courier-guideline/`](.agents/skills/courier-guideline/)：面向 agent 的 courier 使用手册，说明契约定义、routes 组装、client 调用、校验与生成标签等推荐用法。

## 文档导航

- [`docs/index.md`](docs/index.md)：文档首页
- [`docs/quick-start.md`](docs/quick-start.md)：最小可运行示例与入门路径
- [`docs/example-guide.md`](docs/example-guide.md)：`internal/example` 阅读导航
- [`docs/codegen.md`](docs/codegen.md)：生成器职责与使用时机
- [`docs/concepts.md`](docs/concepts.md)：核心概念与主路径
- [`docs/extensions.md`](docs/extensions.md)：关键扩展点说明
- [`docs/boundaries.md`](docs/boundaries.md)：手写、约定与生成的边界
- [`docs/troubleshooting.md`](docs/troubleshooting.md)：常见问题排查
- [`docs/contributing-architecture.md`](docs/contributing-architecture.md)：面向贡献者的架构说明

## 最小入口

- 想先理解最小闭环：看 [`docs/quick-start.md`](docs/quick-start.md)。
- 想直接看真实分层：看 [`internal/example/`](internal/example/) 和 [`docs/example-guide.md`](docs/example-guide.md)。
- 想运行示例服务：执行 `go run ./internal/example/cmd/example`。
