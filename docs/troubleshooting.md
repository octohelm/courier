# 常见问题排查

这份文档聚焦第一次接入 `courier` 时最容易踩的坑。每个问题都尽量给出“现象、原因、排查路径”。

## 1. 改了注解或声明，但生成结果没更新

现象：

- `zz_generated.*.go` 没有同步变化
- 新增的 operator 没有自动注册

常见原因：

- 还没执行 `go generate`
- 改动的目录没有 `//go:generate go tool gen .`

排查路径：

- 看 [`tool/internal/cmd/gen`](https://github.com/octohelm/courier/tree/main/tool/internal/cmd/gen)
- 看 [`devpkg/operatorgen/doc.go`](https://github.com/octohelm/courier/blob/main/devpkg/operatorgen/doc.go)
- 检查目标目录下的 `doc.go`

## 2. OpenAPI 里缺少预期字段

现象：

- 请求参数或响应 schema 没出现在 OpenAPI 中

常见原因：

- 参数字段缺少 `in`、`name` 等必要 tag
- 相关类型没有按 scanner 预期暴露

排查路径：

- 看 [`pkg/courierhttp/openapi`](https://github.com/octohelm/courier/tree/main/pkg/courierhttp/openapi)
- 对照 example 中的 [`internal/example/pkg/endpoints`](https://github.com/octohelm/courier/tree/main/internal/example/pkg/endpoints)

## 3. 请求参数没有按预期绑定

现象：

- path、query 或 body 字段为空
- 结构体字段没有被正确反序列化

常见原因：

- `in:"path"`、`in:"query"`、`in:"body"` 写错或漏写
- path 参数名和路由里的占位符不一致

排查路径：

- 看 [`pkg/courierhttp/transport`](https://github.com/octohelm/courier/tree/main/pkg/courierhttp/transport)
- 看 [`pkg/content`](https://github.com/octohelm/courier/tree/main/pkg/content)
- 对照 [`internal/example/pkg/endpoints/org/v1/orgs.go`](https://github.com/octohelm/courier/blob/main/internal/example/pkg/endpoints/org/v1/orgs.go)

## 4. client 解码结果不符合预期

现象：

- 响应无法反序列化到目标对象
- 错误响应没有按预期解码

常见原因：

- `Content-Type` 与预期 transformer 不匹配
- 调用方传入的目标对象类型不合适

排查路径：

- 看 [`pkg/courierhttp/client/client.go`](https://github.com/octohelm/courier/blob/main/pkg/courierhttp/client/client.go)
- 看 [`pkg/content/transformers`](https://github.com/octohelm/courier/tree/main/pkg/content/transformers)

## 5. route 能跑，但依赖没有注入进来

现象：

- operator 里拿到的 service 为空
- 初始化时报找不到 provider

常见原因：

- provider 没有生成对应的上下文辅助代码
- middleware 没把 service 注入请求上下文

排查路径：

- 看 [`devpkg/injectablegen/doc.go`](https://github.com/octohelm/courier/blob/main/devpkg/injectablegen/doc.go)
- 看 [`internal/example/cmd/example/main.go`](https://github.com/octohelm/courier/blob/main/internal/example/cmd/example/main.go)
- 看 [`internal/example/domain/org/service/service.go`](https://github.com/octohelm/courier/blob/main/internal/example/domain/org/service/service.go)

## 6. 不确定最终到底注册了哪些 HTTP 路由

现象：

- 代码里看起来已经挂了 route，但访问时是 404
- 不确定 `httprouter.New(...)` 最终暴露了哪些 method/path
- 改动前后很难快速确认路由暴露面有没有变化

常见原因：

- route 没有挂到最终传给 `httprouter.New(...)` 的根 router 上
- 中间层 `GroupRouter(...)` 的前缀拼接和预期不一致
- 误以为某个 route 已显式注册，实际上只注册了部分分组

排查路径：

- 先看 [`pkg/courierhttp/handler/httprouter/router.go`](https://github.com/octohelm/courier/blob/main/pkg/courierhttp/handler/httprouter/router.go) 里的 `RouteSnapshot(...)`
- 在启动前直接打印快照，确认最终 method/path/operator 链

```go
snapshot, err := httprouter.RouteSnapshot(routes.R, "example")
if err != nil {
	panic(err)
}

fmt.Println(snapshot)
```

- 再对照 [`internal/example/cmd/example/routes/routes.go`](https://github.com/octohelm/courier/blob/main/internal/example/cmd/example/routes/routes.go) 看根路由组装是否符合预期
- 如果服务已经能启动，也可以直接观察 `httprouter.New(...)` 启动时控制台输出的同款快照

## 7. 校验规则没有按预期生效

现象：

- 输入明显非法，但没有报预期错误

常见原因：

- struct tag 校验声明缺失
- 自定义格式校验器没有注册

排查路径：

- 看 [`pkg/validator`](https://github.com/octohelm/courier/tree/main/pkg/validator)
- 看 [`internal/example/pkg/apis/org/v1/org.go`](https://github.com/octohelm/courier/blob/main/internal/example/pkg/apis/org/v1/org.go)

## 排查顺序建议

遇到问题时，建议先按这个顺序确认：

1. 代码是否已经重新生成
2. endpoint 契约里的 tag 是否正确
3. route 组装层是否挂到了根路由
4. provider 是否已注入 context
5. transformer / client 侧的类型和内容类型是否匹配
6. 最终注册出来的 method/path 是否符合预期
