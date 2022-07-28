package client

import (
	"context"
	"net"
	"net/http"
	"time"

	contextx "github.com/octohelm/x/context"
	"golang.org/x/net/http2"
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

func newDefaultRoundTripper() http.RoundTripper {
	return &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   5 * time.Second,
			KeepAlive: 0,
		}).DialContext,
		DisableKeepAlives:     true,
		TLSHandshakeTimeout:   5 * time.Second,
		ResponseHeaderTimeout: 5 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
}

type RoundTripperCreateFunc = func() http.RoundTripper

type contextRoundTripperCreator struct{}

func ContextWithRoundTripperCreator(ctx context.Context, newRoundTripper RoundTripperCreateFunc) context.Context {
	return contextx.WithValue(ctx, contextRoundTripperCreator{}, newRoundTripper)
}

func RoundTripperCreatorFromContext(ctx context.Context) func() http.RoundTripper {
	if t, ok := ctx.Value(contextRoundTripperCreator{}).(func() http.RoundTripper); ok {
		return t
	}
	return newDefaultRoundTripper
}

func GetShortConnClientContext(ctx context.Context, httpTransports ...HttpTransport) *http.Client {
	t := RoundTripperCreatorFromContext(ctx)()

	if ht, ok := t.(*http.Transport); ok {
		_ = http2.ConfigureTransport(ht)
	}

	for i := range httpTransports {
		t = httpTransports[i](t)
	}

	return &http.Client{Transport: t}
}
