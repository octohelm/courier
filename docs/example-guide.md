# example 阅读导航

`internal/example/` 不是玩具 demo，而是这套库推荐组织方式的真实切片。第一次读时，建议不要按文件名乱跳，而是按层次阅读。

## 推荐阅读顺序

1. 先看 [`internal/example/cmd/example/main.go`](https://github.com/octohelm/courier/blob/main/internal/example/cmd/example/main.go)
2. 再看 [`internal/example/cmd/example/routes/routes.go`](https://github.com/octohelm/courier/blob/main/internal/example/cmd/example/routes/routes.go)
3. 再看 [`internal/example/cmd/example/routes/org/v1`](https://github.com/octohelm/courier/tree/main/internal/example/cmd/example/routes/org/v1) 或 [`internal/example/cmd/example/routes/store/v1`](https://github.com/octohelm/courier/tree/main/internal/example/cmd/example/routes/store/v1)
4. 然后看 [`internal/example/pkg/endpoints`](https://github.com/octohelm/courier/tree/main/internal/example/pkg/endpoints)
5. 最后看 [`internal/example/domain`](https://github.com/octohelm/courier/tree/main/internal/example/domain)

## 分层职责

### `internal/example/pkg/apis`

放领域数据类型、请求体、响应体、错误类型和校验规则。

例如：

- [`internal/example/pkg/apis/org/v1/org.go`](https://github.com/octohelm/courier/blob/main/internal/example/pkg/apis/org/v1/org.go)
- [`internal/example/pkg/apis/store/v1`](https://github.com/octohelm/courier/tree/main/internal/example/pkg/apis/store/v1)

这一层更偏“数据与契约对象”，不负责业务执行。

### `internal/example/pkg/endpoints`

放 endpoint 契约定义，也就是 HTTP 方法、路径、参数位置和响应模型的声明。

例如：

- [`internal/example/pkg/endpoints/org/v1/orgs.go`](https://github.com/octohelm/courier/blob/main/internal/example/pkg/endpoints/org/v1/orgs.go)
- [`internal/example/pkg/endpoints/store/v1`](https://github.com/octohelm/courier/tree/main/internal/example/pkg/endpoints/store/v1)

这一层回答的是：

- 路径是什么
- 参数从 path、query 还是 body 来
- 成功响应和错误响应长什么样

### `internal/example/cmd/example/routes`

放可执行的 route/operator 组装层。

例如：

- [`internal/example/cmd/example/routes/routes.go`](https://github.com/octohelm/courier/blob/main/internal/example/cmd/example/routes/routes.go)
- [`internal/example/cmd/example/routes/org/v1/orgs.go`](https://github.com/octohelm/courier/blob/main/internal/example/cmd/example/routes/org/v1/orgs.go)

这一层会把 endpoint 契约和具体 service 实现接起来。常见形态是：

- 嵌入 endpoint 类型
- 通过 `injectable` 注入 service
- 在 `Output` 中调用 service

### `internal/example/domain`

放业务 service 抽象和内存实现。

例如：

- [`internal/example/domain/org/service`](https://github.com/octohelm/courier/tree/main/internal/example/domain/org/service)
- [`internal/example/domain/store/service`](https://github.com/octohelm/courier/tree/main/internal/example/domain/store/service)

这一层回答的是业务能力如何实现，而不是 HTTP 如何暴露。

## 一次请求如何流动

以组织列表为例：

1. 入口从 [`internal/example/cmd/example/main.go`](https://github.com/octohelm/courier/blob/main/internal/example/cmd/example/main.go) 启动。
2. 根路由在 [`internal/example/cmd/example/routes/routes.go`](https://github.com/octohelm/courier/blob/main/internal/example/cmd/example/routes/routes.go) 把 `/api/example` 挂到组织路由树上。
3. 具体 operator 在 [`internal/example/cmd/example/routes/org/v1/orgs.go`](https://github.com/octohelm/courier/blob/main/internal/example/cmd/example/routes/org/v1/orgs.go)。
4. 它嵌入的 endpoint 契约来自 [`internal/example/pkg/endpoints/org/v1/orgs.go`](https://github.com/octohelm/courier/blob/main/internal/example/pkg/endpoints/org/v1/orgs.go)。
5. 数据类型来自 [`internal/example/pkg/apis/org/v1`](https://github.com/octohelm/courier/tree/main/internal/example/pkg/apis/org/v1)。
6. 实际业务查询委托给 [`internal/example/domain/org/service`](https://github.com/octohelm/courier/tree/main/internal/example/domain/org/service)。

## `pkg/apis` 和 `pkg/endpoints` 的区别

- `pkg/apis` 关心“数据长什么样”。
- `pkg/endpoints` 关心“接口如何暴露这些数据”。

前者偏领域契约，后者偏传输契约。

## `routes` 和 `domain/service` 的区别

- `routes` 负责把 HTTP/operator 世界和业务实现接起来。
- `domain/service` 负责真正的业务逻辑。

如果把业务逻辑直接写进 route，短期会快，但会让契约、传输和实现耦合得更紧。example 的价值就在于把这些层拆开给你看。
