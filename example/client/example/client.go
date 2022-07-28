package example

import (
	"context"

	"github.com/octohelm/courier/pkg/courier"
	"github.com/octohelm/courier/pkg/courierhttp/client"
	_ "github.com/octohelm/courier/pkg/statuserror"
)

// +gengo:client:openapi=http://0.0.0.0:8080/api/example
type Client client.Client

func (c *Client) Do(ctx context.Context, req any, metas ...courier.Metadata) courier.Result {
	return (*client.Client)(c).Do(ctx, req, metas...)
}

func (c *Client) InjectContext(ctx context.Context) context.Context {
	return courier.ContentWithClient(ctx, "example", c)
}
