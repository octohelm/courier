# Endpoint 契约规范

`pkg/endpoints` 应保持为纯契约层：

- 不实现 `Output`
- 不注入具体依赖
- 不注册 router
- 不写存储或业务逻辑

常用写法：

- transport 位置通过 `in:"..."` 声明
- 常见还会配合 `name:"..."`、`mime:"..."`
- 成功返回通过 `ResponseData()` 声明
- 错误集合通过 `ResponseErrors()` 声明

详情不要在 skill 中重复抄写，直接用 `go doc` 查具体包：

```bash
go doc github.com/octohelm/courier/pkg/courierhttp
go doc github.com/octohelm/courier/pkg/courierhttp/transport
go doc github.com/octohelm/courier/pkg/courierhttp/openapi
```

建议：

- 字段说明、接口说明尽量补齐，便于 runtime doc / openapi 生成
- 错误类型在 `pkg/apis/{domain}/{vMajor}` 中定义，并在 endpoint 中显式暴露
- 业务字段的 `validate:"..."` 规则优先定义在 `pkg/apis/{domain}/{vMajor}` 的模型上，而不是在 endpoint 层重复声明
