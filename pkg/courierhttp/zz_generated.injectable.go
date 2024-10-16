/*
Package courierhttp GENERATED BY gengo:injectable 
DON'T EDIT THIS FILE
*/
package courierhttp

import (
	context "context"
)

type contextHttpRequest struct{}

func HttpRequestFromContext(ctx context.Context) (*HttpRequest, bool) {
	if v, ok := ctx.Value(contextHttpRequest{}).(*HttpRequest); ok {
		return v, true
	}
	return nil, false
}

func HttpRequestInjectContext(ctx context.Context, tpe *HttpRequest) context.Context {
	return context.WithValue(ctx, contextHttpRequest{}, tpe)
}

func (p *HttpRequest) InjectContext(ctx context.Context) context.Context {
	return HttpRequestInjectContext(ctx, p)
}

func (v *HttpRequest) Init(ctx context.Context) error {

	return nil
}

type contextOperationInfo struct{}

func OperationInfoFromContext(ctx context.Context) (*OperationInfo, bool) {
	if v, ok := ctx.Value(contextOperationInfo{}).(*OperationInfo); ok {
		return v, true
	}
	return nil, false
}

func OperationInfoInjectContext(ctx context.Context, tpe *OperationInfo) context.Context {
	return context.WithValue(ctx, contextOperationInfo{}, tpe)
}

func (p *OperationInfo) InjectContext(ctx context.Context) context.Context {
	return OperationInfoInjectContext(ctx, p)
}

func (v *OperationInfo) Init(ctx context.Context) error {

	return nil
}

type contextRouteDescriber struct{}

func RouteDescriberFromContext(ctx context.Context) (RouteDescriber, bool) {
	if v, ok := ctx.Value(contextRouteDescriber{}).(RouteDescriber); ok {
		return v, true
	}
	return nil, false
}

func RouteDescriberInjectContext(ctx context.Context, tpe RouteDescriber) context.Context {
	return context.WithValue(ctx, contextRouteDescriber{}, tpe)
}
