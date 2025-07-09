package client

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"
	"time"
)

var reasonableRoundTripper = &http.Transport{
	Proxy: http.ProxyFromEnvironment,

	DialContext: defaultHosts.WrapDialContext((&net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
	}).DialContext),

	MaxIdleConns:        100,
	MaxIdleConnsPerHost: 10,
	IdleConnTimeout:     90 * time.Second,

	TLSHandshakeTimeout:   10 * time.Second,
	ExpectContinueTimeout: 1 * time.Second,

	ResponseHeaderTimeout: 60 * time.Second,

	ForceAttemptHTTP2: true,
}

var defaultTlsConfig = &tls.Config{}

var defaultHosts = Hosts{}

func AddHostAlias(hostAliases ...HostAlias) {
	for _, hostAlias := range hostAliases {
		defaultHosts.AddHostAlias(hostAlias)
	}
}

func SetDefaultTLSClientConfig(tlsConfig *tls.Config) {
	if tlsConfig != nil {
		defaultTlsConfig = tlsConfig.Clone()
		reasonableRoundTripper.TLSClientConfig = tlsConfig.Clone()
	}
}

func GetReasonableClientContext(ctx context.Context, httpTransports ...HttpTransport) *http.Client {
	t := http.RoundTripper(reasonableRoundTripper)

	tc, ok := RoundTripperCreatorFromContext(ctx)
	if ok {
		t = tc()
	}

	return &http.Client{Transport: WithHttpTransports(httpTransports...)(t)}
}

func newRoundTripperWithoutKeepAlive() http.RoundTripper {
	t := &http.Transport{
		Proxy: http.ProxyFromEnvironment,

		DialContext: defaultHosts.WrapDialContext((&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 0,
		}).DialContext),

		DisableKeepAlives: true,

		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		ResponseHeaderTimeout: 60 * time.Second,
	}

	if defaultTlsConfig != nil {
		t.TLSClientConfig = defaultTlsConfig.Clone()
	}

	return t
}

func GetShortConnClientContext(ctx context.Context, httpTransports ...HttpTransport) *http.Client {
	t := newRoundTripperWithoutKeepAlive()

	tc, ok := RoundTripperCreatorFromContext(ctx)
	if ok {
		t = tc()
	}

	return &http.Client{Transport: WithHttpTransports(httpTransports...)(t)}
}
