package openapi

import "context"

type contextKey struct{}

func FromContext(ctx context.Context) *OpenAPI {
	if v, ok := ctx.Value(contextKey{}).(*OpenAPI); ok {
		return v
	}
	return &OpenAPI{}
}

func InjectContext(ctx context.Context, o *OpenAPI) context.Context {
	return context.WithValue(ctx, contextKey{}, o)
}
