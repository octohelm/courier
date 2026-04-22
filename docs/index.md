# courier 文档

这组文档服务于两个目标：

- 帮第一次接触 `courier` 的人快速走通最小闭环。
- 帮准备长期使用 `courier` 的团队理解 example 分层和 codegen 边界。

## 推荐阅读顺序

1. [快速开始](quick-start.md)
2. [example 阅读导航](example-guide.md)
3. [codegen 使用时机](codegen.md)
4. [核心概念与主路径](concepts.md)
5. [关键扩展点说明](extensions.md)
6. [手写、约定与生成的边界](boundaries.md)
7. [常见问题排查](troubleshooting.md)
8. [面向贡献者的架构说明](contributing-architecture.md)

## 文档说明

- `quick-start.md`：适合第一次接触仓库时阅读。
- `example-guide.md`：适合理解 `internal/example/` 的推荐目录结构。
- `codegen.md`：适合在准备引入 `devpkg/` 生成器前阅读。
- `concepts.md`：适合理解整个库的主模型和执行主路径。
- `extensions.md`：适合理解小接口和扩展点为什么存在。
- `boundaries.md`：适合团队决定哪些部分手写、哪些交给约定或生成。
- `troubleshooting.md`：适合排查初次接入时的常见问题。
- `contributing-architecture.md`：适合贡献者评估改动影响面。
