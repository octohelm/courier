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

type contextOperationInfo struct{}

func ContextWithOperationInfo(ctx context.Context, info OperationInfo) context.Context {
	return contextx.WithValue(ctx, contextOperationInfo{}, info)
}

func OperationInfoFromContext(ctx context.Context) OperationInfo {
	if info, ok := ctx.Value(contextOperationInfo{}).(OperationInfo); ok {
		return info
	}
	return OperationInfo{}
}

// OperationIDFromContext
// Deprecated use OperationInfoFromContext instead
func OperationIDFromContext(ctx context.Context) string {
	if info, ok := ctx.Value(contextOperationInfo{}).(OperationInfo); ok {
		return info.ID
	}
	return ""
}

type OperationInfo struct {
	ID     string
	Method string
	Route  string
}
