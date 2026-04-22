# 面向贡献者的架构说明

这篇文档面向准备修改仓库实现的人。目标不是解释所有细节，而是帮助你快速判断“改这块会影响哪里”。

## 目录边界

### `pkg/`

放对外公共库，是仓库最主要的稳定接口面。

重点子目录：

- [`pkg/courier`](https://github.com/octohelm/courier/tree/main/pkg/courier)：最小抽象层
- [`pkg/courierhttp`](https://github.com/octohelm/courier/tree/main/pkg/courierhttp)：HTTP 承载
- [`pkg/openapi`](https://github.com/octohelm/courier/tree/main/pkg/openapi)：OpenAPI 文档对象
- [`pkg/validator`](https://github.com/octohelm/courier/tree/main/pkg/validator)：校验能力

### `internal/`

放仓库内部支撑实现和 example。

重点子目录：

- [`internal/request`](https://github.com/octohelm/courier/tree/main/internal/request)：route handler 组装
- [`internal/pathpattern`](https://github.com/octohelm/courier/tree/main/internal/pathpattern)：路径匹配结构
- [`internal/example`](https://github.com/octohelm/courier/tree/main/internal/example)：完整示例

### `devpkg/`

放开发期生成器和生成辅助逻辑。

适合在这些场景下查看：

- operator 注册或响应推导问题
- injectable 上下文辅助代码问题
- OpenAPI 到 client 代码生成问题

### `tool/internal/cmd/gen`

仓库统一的生成入口。改动生成链路时通常需要联动检查这里。

## 改动影响面速查

### 如果你改 `pkg/courier`

同步检查：

- `pkg/courierhttp`
- `internal/request`
- `pkg/courier` 自身测试

原因：

- 这里是主抽象层，任何生命周期或 route 行为变化都会向上游扩散。

### 如果你改 `pkg/courierhttp/transport` 或 `pkg/content`

同步检查：

- `pkg/courierhttp/client`
- `pkg/courierhttp/handler`
- `pkg/courierhttp/openapi`

原因：

- 请求绑定、响应写回和 client 解码都依赖这里。

### 如果你改 `pkg/courierhttp/openapi` 或 `pkg/openapi`

同步检查：

- example 中的 endpoint 定义
- OpenAPI 相关测试
- 可能依赖 OpenAPI 的 client 生成逻辑

### 如果你改 `devpkg/*`

同步检查：

- 相关 `doc.go`
- example 里的生成结果
- `go generate` 是否仍能顺利执行

## 推荐改动顺序

1. 先确定改动落在哪个边界层
2. 再确认上游或下游依赖面
3. 先补或调整测试
4. 再修改实现
5. 最后检查 example 和生成结果是否仍然成立

## 为什么要多看 example

很多仓库的 example 只是展示 API，但这里的 [`internal/example`](https://github.com/octohelm/courier/tree/main/internal/example) 更接近“集成测试级别的参考样本”。准备改公共行为时，先对照 example，通常能更快发现真实使用面。
