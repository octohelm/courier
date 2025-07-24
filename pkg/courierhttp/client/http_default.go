package client

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"
	"time"

	"golang.org/x/net/http2"
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

func UpgradeToSupportH2c(t http.RoundTripper) http.RoundTripper {
	if t1, ok := t.(*http.Transport); ok {
		if !t1.DisableKeepAlives {
			if t2, err := http2.ConfigureTransports(t1); err == nil {
				t2.AllowHTTP = true
				t2.DialTLSContext = func(ctx context.Context, network, addr string, cfg *tls.Config) (net.Conn, error) {
					return t1.DialContext(ctx, network, addr)
				}
				t2.ConnPool = nil

				return t2
			}
		}
	}

	return t
}
