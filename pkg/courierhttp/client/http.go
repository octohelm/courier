package client

import (
	"context"
	"net/http"

	contextx "github.com/octohelm/x/context"
)

type contextKeyClient struct{}

func ContextWithHttpClient(ctx context.Context, c *http.Client) context.Context {
	return contextx.WithValue(ctx, contextKeyClient{}, c)
}

func HttpClientFromContext(ctx context.Context) *http.Client {
	if c, ok := ctx.Value(contextKeyClient{}).(*http.Client); ok {
		return c
	}
	return nil
}

type RoundTripperCreateFunc = func() http.RoundTripper

type contextRoundTripperCreator struct{}

func ContextWithRoundTripperCreator(ctx context.Context, newRoundTripper RoundTripperCreateFunc) context.Context {
	return contextx.WithValue(ctx, contextRoundTripperCreator{}, newRoundTripper)
}

func RoundTripperCreatorFromContext(ctx context.Context) (func() http.RoundTripper, bool) {
	if t, ok := ctx.Value(contextRoundTripperCreator{}).(func() http.RoundTripper); ok {
		return t, true
	}
	return nil, false
}
