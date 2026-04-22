# codegen 使用时机

`courier` 不是“所有东西都必须生成”的库，但它确实把 codegen 作为降低样板代码的重要手段。判断是否该生成，建议先看“你想省掉哪一类重复代码”。

## 仓库里的生成器

### `devpkg/operatorgen`

作用：

- 为 operator 生成路由注册代码
- 推断响应内容和状态码
- 扫描错误返回并补 `ResponseErrors`

入口说明见 [`devpkg/operatorgen/doc.go`](https://github.com/octohelm/courier/blob/main/devpkg/operatorgen/doc.go)。

典型场景：

- 你已经有导出的 operator 类型
- 想自动注册到某个 router 变量
- 不想手写 `ResponseContent`、`ResponseErrors` 一类样板

example 中的生成结果可参考：

- [`internal/example/cmd/example/routes/org/v1/zz_generated.operator.go`](https://github.com/octohelm/courier/blob/main/internal/example/cmd/example/routes/org/v1/zz_generated.operator.go)
- [`internal/example/cmd/example/routes/store/v1/zz_generated.operator.go`](https://github.com/octohelm/courier/blob/main/internal/example/cmd/example/routes/store/v1/zz_generated.operator.go)

### `devpkg/clientgen`

作用：

- 根据 OpenAPI 文档生成客户端操作类型与相关 schema

入口说明见 [`devpkg/clientgen/doc.go`](https://github.com/octohelm/courier/blob/main/devpkg/clientgen/doc.go)。

典型场景：

- 你已经有稳定的 OpenAPI 文档
- 想生成一组 typed client，而不是手写请求模型

### `devpkg/injectablegen`

作用：

- 为 provider 和可注入 struct 生成上下文注入与初始化辅助代码

入口说明见 [`devpkg/injectablegen/doc.go`](https://github.com/octohelm/courier/blob/main/devpkg/injectablegen/doc.go)。

典型场景：

- 你希望 route/operator 通过 context 注入 service
- 不想手写 `FromContext` / `InjectContext` / `Init`

### `devpkg/uintstrgen`

作用：

- 为无符号整数别名生成文本编解码和字符串表示方法

入口说明见 [`devpkg/uintstrgen/doc.go`](https://github.com/octohelm/courier/blob/main/devpkg/uintstrgen/doc.go)。

典型场景：

- 你有类似 `OrgID uint64` 这样的类型
- 想让它天然支持文本编解码和字符串表示

example 可参考 [`internal/example/pkg/apis/org/v1/org.go`](https://github.com/octohelm/courier/blob/main/internal/example/pkg/apis/org/v1/org.go) 中的 `OrgID`。

## 什么时候需要 `go generate`

当你改动了依赖生成器的声明时，通常需要重新生成。例如：

- 新增或修改了 `+gengo:operator:*`
- 新增或修改了 `+gengo:injectable*`
- 新增了 `+gengo:uintstr`
- 新增了基于 OpenAPI 的 client 声明

仓库里统一通过 [`tool/internal/cmd/gen`](https://github.com/octohelm/courier/tree/main/tool/internal/cmd/gen) 暴露 `go tool gen` 入口。根 `justfile` 也提供了：

```bash
just gen ./...
```

如果只想针对某个目录执行，也可以直接：

```bash
go generate ./internal/example/...
```

## 什么时候可以先手写

不是所有场景都必须生成。下面这些场景可以先手写：

- 你只是在验证某个 endpoint 契约是否合理
- 你还没决定 router 注册方式
- 你只需要极少量样板代码，手写更清楚

更实用的经验是：

- 样板开始重复时，再引入生成器
- 团队需要统一风格时，再把生成器纳入默认流程

## 怎么判断要不要生成

可以用这个简单判断：

- 如果你在重复写注册、上下文注入、响应描述，优先考虑生成。
- 如果你还在快速试验接口形状，先手写通常更直观。
- 如果你已经依赖 OpenAPI 作为契约源，优先考虑 `clientgen`。
