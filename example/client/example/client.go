package example

import (
	"context"

	contextx "github.com/octohelm/x/context"

	"github.com/octohelm/courier/pkg/courier"
	"github.com/octohelm/courier/pkg/courierhttp/client"

	_ "io"

	_ "github.com/octohelm/courier/example/apis/org"
	_ "github.com/octohelm/courier/example/pkg/domain/org"
	_ "github.com/octohelm/courier/example/pkg/filter"
	_ "github.com/octohelm/courier/pkg/statuserror"
)

// +gengo:client:openapi=http://0.0.0.0:9001/api/example
type Client client.Client

func (c *Client) Do(ctx context.Context, req any, metas ...courier.Metadata) courier.Result {
	return (*client.Client)(c).Do(ctx, req, metas...)
}

func (c *Client) InjectContext(ctx context.Context) context.Context {
	return ClientContext.Inject(ctx, c)
}

var ClientContext = contextx.New[*Client]()

func Do[Data any, Op interface{ ResponseData() *Data }](ctx context.Context, req Op, metas ...courier.Metadata) (*Data, error) {
	return courier.DoWith[Data, Op](ctx, ClientContext.From(ctx), req, metas...)
}
