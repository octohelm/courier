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

func ClientFromContent(ctx context.Context, name string) Client {
	if v, ok := ctx.Value(clientContext{name: name}).(Client); ok {
		return v
	}
	return nil
}

func ContentWithClient(ctx context.Context, name string, client Client) context.Context {
	return contextx.WithValue(ctx, clientContext{name: name}, client)
}
