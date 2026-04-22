# 快速开始

这篇文档的目标不是完整解释 `courier`，而是帮助你先建立一个最小闭环：这个库的核心对象是谁，示例服务怎么跑起来，路由和 OpenAPI 是怎么接上的。

## 先理解三个角色

### 1. Operator

`courier` 的核心动作单位是 `Operator`。你可以把它理解成“一个可执行的接口操作定义”，最核心的约束在 [`pkg/courier/operator.go`](https://github.com/octohelm/courier/blob/main/pkg/courier/operator.go)：

- 输入通常来自 struct 字段和 tag。
- 输出统一走 `Output(context.Context) (any, error)`。

### 2. Router

`Router` 负责把多个 operator 组织成执行链。核心实现见 [`pkg/courier/router.go`](https://github.com/octohelm/courier/blob/main/pkg/courier/router.go)。

在 example 里，根路由组装在 [`internal/example/cmd/example/routes/routes.go`](https://github.com/octohelm/courier/blob/main/internal/example/cmd/example/routes/routes.go)：

- `/api/example` 挂组织相关路由
- `/api/store` 挂存储相关路由
- `httprouter.OpenAPI` 负责补 OpenAPI 输出入口

### 3. HTTP 入口

HTTP 服务启动入口在 [`internal/example/cmd/example/main.go`](https://github.com/octohelm/courier/blob/main/internal/example/cmd/example/main.go)：

- `httprouter.New(...)` 把 `courier.Router` 转成 `http.Handler`
- `handler.ApplyMiddlewares(...)` 注入 example 里的 service
- `httputil.ListenAndServe(...)` 启动监听

## 跑起仓库自带示例

在仓库根目录执行：

```bash
go run ./internal/example/cmd/example
```

服务默认监听：

```text
http://127.0.0.1:9001
```

如果你只想在 example 目录里运行，也可以进入 [`internal/example/cmd/example`](https://github.com/octohelm/courier/tree/main/internal/example/cmd/example) 后执行：

```bash
just serve
```

## 试一个请求

example 里组织列表的 endpoint 定义在 [`internal/example/pkg/endpoints/org/v1/orgs.go`](https://github.com/octohelm/courier/blob/main/internal/example/pkg/endpoints/org/v1/orgs.go)。

服务启动后，可以先请求：

```bash
curl 'http://127.0.0.1:9001/api/example/v1/orgs'
```

这个请求会经过：

1. `routes.R` 中的 `/api` 和 `/example` 分组
2. `ListOrg` 对应的 endpoint/operator
3. service 注入后的内存实现
4. `courierhttp` 的响应写回流程

## 看 OpenAPI

`httprouter.New(...)` 会在未显式注册自定义 OpenAPI 路由时，补齐默认的 `OpenAPI` 和 `OpenAPIView` 路由。对应类型在：

- [`pkg/courierhttp/handler/httprouter/openapi.go`](https://github.com/octohelm/courier/blob/main/pkg/courierhttp/handler/httprouter/openapi.go)
- [`pkg/courierhttp/handler/httprouter/openapi_view.go`](https://github.com/octohelm/courier/blob/main/pkg/courierhttp/handler/httprouter/openapi_view.go)

example 本身已经在 [`internal/example/cmd/example/routes/routes.go`](https://github.com/octohelm/courier/blob/main/internal/example/cmd/example/routes/routes.go) 里显式注册了 `&httprouter.OpenAPI{}`，所以这里不会再额外自动补一个默认 `OpenAPI` 路由。只有未显式注册时，`httprouter.New(...)` 才会补齐默认的 `OpenAPI` 和 `OpenAPIView` 路由。

## 下一步看什么

- 想理解 example 的目录为什么这样分：继续看 [example 阅读导航](example-guide.md)
- 想知道哪些代码建议交给生成器：继续看 [codegen 使用时机](codegen.md)
