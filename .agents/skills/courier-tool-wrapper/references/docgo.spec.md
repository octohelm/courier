# `doc.go` 包级标签规范

常用的全局 `gengo` 标签应放在包级 `doc.go` 中，而不是散落到普通业务文件。

## 常见约定

契约层包通常开启 runtime doc：

```go
// +gengo:runtimedoc=true
//
//go:generate go tool gen .
package v1
```

典型位置：

- `pkg/apis/{domain}/{vMajor}/doc.go`
- `pkg/endpoints/{domain}/{vMajor}/doc.go`

`cmd/{app}/routes/{domain}/{vMajor}` 通常负责 operator 注册，并关闭 runtime doc：

```go
// +gengo:operator:register=R
// +gengo:runtimedoc=false
//
//go:generate go tool gen .
package v1

import "github.com/octohelm/courier/pkg/courier"

var R = courier.NewRouter()
```

## 建议

- 包级生成标签统一放在 `doc.go`
- `pkg/apis`、`pkg/endpoints` 默认作为文档与契约来源，优先开启 `+gengo:runtimedoc=true`
- `cmd/*/routes` 作为组装层，常见做法是通过 `+gengo:operator:register=R` 注册 router，并关闭 runtime doc
- `R` 这类 router 入口保持在包级，便于生成注册代码对齐

## 查询指引

具体生成结果和标签行为，优先回到项目中的真实 `doc.go` 与生成产物对照理解。

如果需要确认 router/operator 相关能力，直接查：

```bash
go doc github.com/octohelm/courier/pkg/courier
```
