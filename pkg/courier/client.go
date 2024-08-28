package courier

import (
	"context"

	contextx "github.com/octohelm/x/context"
)

type Client interface {
	Do(ctx context.Context, req any, metas ...Metadata) Result
}

type Result interface {
	Into(v any) (Metadata, error)
}

type clientContext struct{ name string }

func ClientFromContext(ctx context.Context, name string) Client {
	if v, ok := ctx.Value(clientContext{name: name}).(Client); ok {
		return v
	}
	return nil
}

func ContextWithClient(ctx context.Context, name string, client Client) context.Context {
	return contextx.WithValue(ctx, clientContext{name: name}, client)
}

func DoWith[Data any, Op interface{ ResponseData() *Data }](ctx context.Context, c Client, req Op, metas ...Metadata) (*Data, error) {
	resp := new(Data)

	if _, ok := any(resp).(*NoContent); ok {
		_, err := c.Do(ctx, req, metas...).Into(nil)
		return resp, err
	}

	_, err := c.Do(ctx, req, metas...).Into(resp)
	return resp, err
}

type NoContent struct {
}
