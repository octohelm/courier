# +gengo:injectable 规范

推荐把“可注入依赖”显式声明出来，再让生成器补齐注入辅助代码。

常用模式：

- 在可注入依赖类型上标注 `// +gengo:injectable:provider`
- 在需要被注入的 wrapper struct 上标注 `// +gengo:injectable`
- 通过 `inject:""` 声明字段注入点
- `zz_generated.injectable.go` 视为生成产物，不手写、不手改

可注入依赖既可以是 interface，也可以是 struct。

- 需要隔离实现、替换 mock、跨模块复用时，可以定义 interface
- 对于常规简单业务，直接注入 struct 也可以，不必为了注入而先抽象一层接口
- 需要一次暴露多个 provider 时，也可以使用聚合 struct

interface 示例：

```go
import (
	"context"

	orgv1 "github.com/octohelm/courier/pkg/apis/org/v1"
)

// +gengo:injectable:provider
type OrgService interface {
	Create(ctx context.Context, req *orgv1.OrgForCreateRequest) (*orgv1.Org, error)
}
```

struct 示例：

```go
import (
	"context"

	orgv1 "github.com/octohelm/courier/pkg/apis/org/v1"
)

// +gengo:injectable:provider
type OrgService struct{}

func (OrgService) Create(ctx context.Context, req *orgv1.OrgForCreateRequest) (*orgv1.Org, error) {
	return nil, nil
}
```

聚合 provider 示例：

```go
// +gengo:injectable:provider
type XProvider struct {
	A X `provide:""`
	B B `provide:""`
}
```

wrapper 示例：

```go
import (
	"context"

	endpointorgv1 "github.com/octohelm/courier/pkg/endpoints/org/v1"
)

// +gengo:injectable
type CreateOrg struct {
	endpointorgv1.CreateOrg

	svc OrgService `inject:""`
}

func (r *CreateOrg) Output(ctx context.Context) (any, error) {
	return r.svc.Create(ctx, &r.Body)
}
```

关键点：

- routes 通过注入依赖连接实现，不把构造逻辑揉进 operator
- 入口可以注入单例、mock、mem、struct service 或其他实现
- provider 可以是单个可注入类型，也可以是聚合多个 provider 的 struct
- 具体实现目录不固定，只要能暴露稳定 provider 即可

如果需要确认生成侧或注入侧具体能力，优先回到项目里的真实 provider 定义和生成产物，再用 `go doc` 查对应包。
