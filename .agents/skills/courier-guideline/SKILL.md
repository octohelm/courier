---
name: courier-guideline
description: 当任务涉及基于 github.com/octohelm/courier 定义 API 契约、组装 routes 或编写 client 调用时使用。
---

# Courier Guideline

按 `github.com/octohelm/courier` 约定定义 API——契约、endpoint、组装。

## 定义一个 API

按 `apis → endpoints → routes` 三层递进：

**1. apis（模型层）**：
```go
// pkg/apis/{domain}/v1/org.go
type Org struct {
    ID   OrgID  `json:"id"`
    Name string `json:"name"`
}
```

**2. endpoints（HTTP 契约）**：
```go
// pkg/endpoints/{domain}/v1/orgs.go
import "github.com/octohelm/courier/pkg/courierhttp"

type CreateOrg struct {
    courierhttp.MethodPost `path:"/orgs"`
    Body CreateOrgRequest  `in:"body"`
}
func (CreateOrg) ResponseData() *Org { return new(Org) }
func (CreateOrg) ResponseErrors() []error { return []error{&ErrOrgNameConflict{}} }
```

**3. routes（组装 + 注入）**：
```go
// cmd/{app}/routes/{domain}/v1/orgs.go
// +gengo:injectable  ← gengo 标签，需要 gengo-guideline
type CreateOrg struct {
    endpointorgv1.CreateOrg
    svc OrgService `inject:""`
}
func (r *CreateOrg) Output(ctx context.Context) (any, error) {
    return r.svc.Create(ctx, &r.Body)
}
```

routes 层的 `+gengo:injectable`、endpoints 的 `+gengo:runtimedoc` 等注解由 gengo 引擎驱动——生成器注册和注解语法参考 gengo-guideline。

## 更多

- 校验和自定义类型 → [references/validation.spec.md](references/validation.spec.md)
- 错误定义 → [references/errors.spec.md](references/errors.spec.md)
- 注入依赖 → [references/injectable.spec.md](references/injectable.spec.md)
- 分层与命名 → [references/layout.spec.md](references/layout.spec.md)
- gengo 标签和生成器 → [references/docgo.spec.md](references/docgo.spec.md) + gengo-guideline
- client / 测试 → [references/testing.md](references/testing.md)

API 细节以 `go doc github.com/octohelm/courier/pkg/courier` 和 `go doc github.com/octohelm/courier/pkg/courierhttp` 为准。
