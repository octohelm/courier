# Endpoint 契约规范

> 以下涉及的目录路径（`pkg/endpoints`、`pkg/apis/...`）是使用者项目约定，非 courier 自身包。

`pkg/endpoints` 是纯契约层：不实现业务逻辑、不注入依赖、不注册 router。

## 写法

```go
import "github.com/octohelm/courier/pkg/courierhttp"

// 创建组织
type CreateOrg struct {
	courierhttp.MethodPost `path:"/orgs"`
	Body CreateOrgRequest  `in:"body"`
}

func (CreateOrg) ResponseData() *Org {
	return new(Org)
}

func (CreateOrg) ResponseErrors() []error {
	return []error{&ErrOrgNameConflict{}}
}
```

- 嵌入 `courierhttp.MethodPost` / `MethodGet` / `MethodPut` / `MethodDelete` / `MethodPatch` 声明 HTTP 方法
- `path:"..."` 声明路由，参数用 `{name}` 占位
- 字段用 `in:"body"` / `in:"path"` / `in:"query"` / `in:"header"` 声明参数位置，可配合 `name:"..."` 和 `mime:"..."`
- `ResponseData()` 声明成功返回类型
- `ResponseErrors()` 声明可能返回的错误

具体类型和接口细节通过 `go doc` 查阅：
- `go doc github.com/octohelm/courier/pkg/courierhttp`
- `go doc github.com/octohelm/courier/pkg/courierhttp/transport`
- `go doc github.com/octohelm/courier/pkg/courierhttp/openapi`
