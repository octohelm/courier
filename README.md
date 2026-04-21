# courier

[![GoDoc Widget](https://godoc.org/github.com/octohelm/courier?status.svg)](https://pkg.go.dev/github.com/octohelm/courier)

`courier` 是一个基于 Go 的服务框架基础库，围绕接口建模、HTTP 路由承载、OpenAPI 描述生成，以及配套的校验与错误表达提供可复用能力。

仓库同时包含核心库实现、代码生成辅助包和可运行示例，适合作为服务框架底座或上层业务框架的依赖。

## 内容总览

- [`pkg/`](pkg/)：核心公共库，包含 `courier` 抽象、HTTP 承载、OpenAPI、校验、状态错误与表达式等能力。
- [`devpkg/`](devpkg/)：面向开发期的生成器与辅助包，用于支撑 operator、client、injectable 等代码生成。
- [`internal/example/`](internal/example/)：仓库内示例实现，展示 `pkg/{apis,endpoints}` 契约层、`cmd/example/routes` 组装层与注入式 service 的组织方式。
- [`.agents/skills/courier-tool-wrapper/`](.agents/skills/courier-tool-wrapper/)：面向 agent 的 `courier` 使用手册，说明契约定义、routes 组装、client 调用、校验与生成标签等推荐用法。

## 相关文档

- [`AGENTS.md`](AGENTS.md)：仓库级协同约束、变更边界与人工接管条件。
- [`justfile`](justfile)：仓库级常用命令入口，覆盖生成、格式化、测试与示例运行。
