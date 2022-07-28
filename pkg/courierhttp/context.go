package courierhttp

import (
	"context"
	"net/http"

	contextx "github.com/octohelm/x/context"
)

type contextKeyHttpRequestKey struct{}

func ContextWithHttpRequest(ctx context.Context, req *http.Request) context.Context {
	return contextx.WithValue(ctx, contextKeyHttpRequestKey{}, req)
}

func HttpRequestFromContext(ctx context.Context) *http.Request {
	p, _ := ctx.Value(contextKeyHttpRequestKey{}).(*http.Request)
	return p
}

type contextKeyOperationID struct{}

func ContextWithOperationID(ctx context.Context, operationID string) context.Context {
	return contextx.WithValue(ctx, contextKeyOperationID{}, operationID)
}

func OperationIDFromContext(ctx context.Context) string {
	if id, ok := ctx.Value(contextKeyOperationID{}).(string); ok {
		return id
	}
	return ""
}
