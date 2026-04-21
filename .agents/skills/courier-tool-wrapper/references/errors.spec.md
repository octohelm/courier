# Domain Errors 规范

domain errors 属于契约层，和领域模型、请求体一样，应放在版本化 API 包中定义。

推荐位置：

- `pkg/apis/{domain}/{vMajor}/errors.go`

推荐原则：

- 一个 domain 的错误类型放在对应 domain 的版本目录下
- 错误类型名称直接表达业务语义，例如 `ErrOrgNotFound`、`ErrOrgNameConflict`
- 错误语义面向调用方，而不是面向内部实现细节
- 字段说明与错误说明尽量补齐，便于文档与 client 理解

endpoint 暴露原则：

- endpoint 通过 `ResponseErrors() []error` 显式声明可能返回的 domain errors
- `ResponseErrors()` 只暴露该接口真实可能出现的错误集合
- 不要把内部实现细节错误直接泄漏为契约错误，必要时先转换为 domain errors

组织建议：

- 按 domain 聚合错误，而不是按 handler 文件拆散错误
- 错误定义放在 `pkg/apis/{domain}/{vMajor}`，不要放进 `pkg/endpoints`
- 如果同一错误会被多个 endpoint 复用，仍保持在同一个 domain errors 文件中

命名建议：

- 优先使用 `Err{Domain}{Reason}` 形式
- `Reason` 应直接体现业务失败原因，如 `NotFound`、`Conflict`、`Invalid`
- 避免把 transport、存储或第三方依赖名字带入公开错误类型
