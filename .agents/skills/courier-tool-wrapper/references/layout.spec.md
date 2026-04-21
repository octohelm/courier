# 分层与命名规范

推荐职责拆分：

- `pkg/{apis,endpoints}`
  作为契约层。
- `pkg/apis/{domain}/{vMajor}`
  放领域数据结构、枚举、请求体、列表结构、错误类型、字段说明、validate 规则与自定义文本类型。
- `pkg/endpoints/{domain}/{vMajor}`
  放接口契约定义。
  每个 endpoint 负责声明 `Method + path + 参数位置 + ResponseData() + ResponseErrors()`。
- `cmd/{app}/routes`
  负责服务组装和 client 调用对接。
- `实现层目录`
  可放在 `domain/*`、`biz/*`、`application/*`、`service/*` 或其他项目约定位置，不强制固定目录。

推荐推进顺序：

1. 先写 `pkg/apis/{domain}/{vMajor}`
2. 再写 `pkg/endpoints/{domain}/{vMajor}`
3. 再定义实现层抽象和实现
4. 最后写 `cmd/*/routes/{domain}/{vMajor}`

命名建议：

- `apis` 文件名按领域对象命名，不按 handler 动词命名
- `endpoints` 文件名按 domain 聚合，例如 `orgs.go`、`blobs.go`、`manifests.go`
- `routes` 下的 wrapper 分组尽量与 endpoint 对齐
- `vMajor` 使用明确版本目录，例如 `v0`、`v1`
