package courier

import (
	"context"

	contextx "github.com/octohelm/x/context"
)

// Client 表示服务客户端接口，用于执行远程调用。
type Client interface {
	Do(ctx context.Context, req any, metas ...Metadata) Result
}

// Result 表示客户端调用结果接口。
type Result interface {
	Into(v any) (Metadata, error)
}

type clientContext struct{ name string }

// ClientFromContext 从上下文中获取指定名称的客户端。
func ClientFromContext(ctx context.Context, name string) Client {
	if v, ok := ctx.Value(clientContext{name: name}).(Client); ok {
		return v
	}
	return nil
}

// ContextWithClient 将客户端存储到上下文中。
func ContextWithClient(ctx context.Context, name string, client Client) context.Context {
	return contextx.WithValue(ctx, clientContext{name: name}, client)
}
