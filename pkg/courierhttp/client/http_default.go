package client

import (
	"context"
	"net"
	"net/http"
	"time"
)

var reasonableRoundTripper http.RoundTripper = &http.Transport{
	Proxy: http.ProxyFromEnvironment,

	DialContext: (&net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
	}).DialContext,

	MaxIdleConns:        100,
	MaxIdleConnsPerHost: 10,
	IdleConnTimeout:     90 * time.Second,

	TLSHandshakeTimeout:   10 * time.Second,
	ExpectContinueTimeout: 1 * time.Second,

	ResponseHeaderTimeout: 60 * time.Second,

	ForceAttemptHTTP2: true,
}

func GetReasonableClientContext(ctx context.Context, httpTransports ...HttpTransport) *http.Client {
	t := reasonableRoundTripper

	tc, ok := RoundTripperCreatorFromContext(ctx)
	if ok {
		t = tc()
	}

	for i := range httpTransports {
		t = httpTransports[i](t)
	}

	return &http.Client{Transport: t}
}

func newRoundTripperWithoutKeepAlive() http.RoundTripper {
	return &http.Transport{
		Proxy: http.ProxyFromEnvironment,

		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 0,
		}).DialContext,

		DisableKeepAlives: true,

		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		ResponseHeaderTimeout: 60 * time.Second,
	}
}

func GetShortConnClientContext(ctx context.Context, httpTransports ...HttpTransport) *http.Client {
	t := newRoundTripperWithoutKeepAlive()

	tc, ok := RoundTripperCreatorFromContext(ctx)
	if ok {
		t = tc()
	}

	for i := range httpTransports {
		t = httpTransports[i](t)
	}

	return &http.Client{Transport: t}
}
