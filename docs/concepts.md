# 核心概念与主路径

`courier` 的包不少，但主模型其实并不分散。理解它时，建议先抓住五个核心角色，再去看具体包。

## 五个核心角色

### 1. Operator

`Operator` 是最核心的执行单元，定义在 [`pkg/courier/operator.go`](https://github.com/octohelm/courier/blob/main/pkg/courier/operator.go)。

你可以把它理解成：

- 一个接口操作的可执行定义
- 一个统一的输入输出边界
- 其他能力的挂载点

最核心的方法只有：

```go
Output(context.Context) (any, error)
```

### 2. Router

`Router` 负责把一个或多个 operator 组织成执行链，定义见 [`pkg/courier/router.go`](https://github.com/octohelm/courier/blob/main/pkg/courier/router.go)。

它的职责不是直接处理 HTTP，而是描述：

- 哪些 operator 会串在一起执行
- 哪些 path group 或 base path 会叠加
- 哪个 operator 是最后一个输出点

### 3. Transport

`Transport` 负责把 `Router` 真正承载到某个传输协议上。最基础的接口在 [`pkg/courier/transport.go`](https://github.com/octohelm/courier/blob/main/pkg/courier/transport.go)。

在这个仓库里，最主要的传输承载是 HTTP，对应实现集中在：

- [`pkg/courierhttp`](https://github.com/octohelm/courier/tree/main/pkg/courierhttp)
- [`pkg/courierhttp/transport`](https://github.com/octohelm/courier/tree/main/pkg/courierhttp/transport)
- [`pkg/courierhttp/handler/httprouter`](https://github.com/octohelm/courier/tree/main/pkg/courierhttp/handler/httprouter)

### 4. OpenAPI

OpenAPI 在这套库里不是额外挂的一份文档，而是从路由和类型信息扫描出来的契约投影。核心能力在：

- [`pkg/courierhttp/openapi`](https://github.com/octohelm/courier/tree/main/pkg/courierhttp/openapi)
- [`pkg/openapi`](https://github.com/octohelm/courier/tree/main/pkg/openapi)

### 5. Client

client 不是旁支，而是复用同一套契约信息做远程调用。核心能力在：

- [`pkg/courierhttp/client`](https://github.com/octohelm/courier/tree/main/pkg/courierhttp/client)
- [`pkg/courierhttp/transport/outgoing_transport.go`](https://github.com/octohelm/courier/blob/main/pkg/courierhttp/transport/outgoing_transport.go)

## 一次请求的主路径

把这套库放到 HTTP 场景里看，一次请求大致会经过下面路径：

1. 先定义 operator 类型和输入输出字段。
2. 再用 `Router` 把 operator 组织成执行链。
3. `courierhttp` 把这个执行链承载到 HTTP handler。
4. 请求进入后，由 transport 和 content 层完成参数绑定与反序列化。
5. operator 执行后，由 `courierhttp` 统一写回响应或错误。
6. OpenAPI scanner 根据同一套路由和类型信息生成文档。
7. client 侧可复用契约信息生成请求和解码响应。

## 为什么这套模型重要

这条主路径的价值不在于“功能多”，而在于这些功能不是彼此割裂的：

- 请求绑定不是另一套模型
- OpenAPI 不是手工维护的第二份契约
- client 也不是纯手搓

这让 `courier` 更像一个服务基础设施底座，而不是单个 HTTP 包。

## 在 example 里对应到哪里

如果想把上面的主路径对应到仓库里的真实代码，可以按下面顺序看：

1. [`internal/example/pkg/endpoints`](https://github.com/octohelm/courier/tree/main/internal/example/pkg/endpoints)
2. [`internal/example/cmd/example/routes`](https://github.com/octohelm/courier/tree/main/internal/example/cmd/example/routes)
3. [`internal/example/cmd/example/main.go`](https://github.com/octohelm/courier/blob/main/internal/example/cmd/example/main.go)
4. [`pkg/courierhttp/handler/httprouter`](https://github.com/octohelm/courier/tree/main/pkg/courierhttp/handler/httprouter)
5. [`pkg/courierhttp/openapi`](https://github.com/octohelm/courier/tree/main/pkg/courierhttp/openapi)

## 读这套库时容易误解的点

- `pkg/courier` 很薄，不代表库本身很轻。真正的体系能力在 `courierhttp`、`openapi`、`validator` 和 `devpkg`。
- `internal/example` 不是附属 demo，而是推荐分层的主要样本。
- 生成器不是可有可无的装饰，而是降低样板代码的重要部分。
