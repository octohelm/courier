package client

import (
	"bytes"
	"context"
	"crypto/tls"
	"errors"
	"io"
	"net"
	"net/http"
	"regexp"
	"testing"

	. "github.com/octohelm/x/testing/v2"
	"golang.org/x/net/http2"

	"github.com/octohelm/courier/pkg/courier"
	"github.com/octohelm/courier/pkg/courierhttp"
)

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) { return f(req) }

type noContentRequest struct{}

func (noContentRequest) ResponseData() *courier.NoContent { return &courier.NoContent{} }

type decodeError struct {
	statusCode int
	body       string
}

func (e *decodeError) Error() string { return e.body }

func (e *decodeError) UnmarshalErrorResponse(statusCode int, respBody []byte) error {
	e.statusCode = statusCode
	e.body = string(respBody)
	return nil
}

type customContentType struct {
	Value string `json:"value"`
}

func (*customContentType) ContentType() string { return "application/json" }

type testRequest struct {
	courierhttp.MethodGet `path:"/users/:id"`
	ID                    string `name:"id" in:"path"`
}

func TestClientContextAndTransportHelpers(t0 *testing.T) {
	httpClient := &http.Client{}
	ctx := ContextWithHttpClient(context.Background(), httpClient)
	ctx = ContextWithRoundTripperCreator(ctx, func() http.RoundTripper {
		return roundTripperFunc(func(req *http.Request) (*http.Response, error) {
			return &http.Response{StatusCode: http.StatusNoContent, Body: io.NopCloser(bytes.NewBuffer(nil)), Header: http.Header{}, Request: req}, nil
		})
	})

	Then(t0, "上下文与 transport 辅助方法符合预期",
		Expect(HttpClientFromContext(ctx), Equal(httpClient)),
		ExpectMust(func() error {
			create, ok := RoundTripperCreatorFromContext(ctx)
			if !ok || create == nil {
				return errClient("missing round tripper creator")
			}
			if _, ok := create().(roundTripperFunc); !ok {
				return errClient("unexpected round tripper type")
			}
			return nil
		}),
		ExpectMust(func() error {
			applied := false
			rt := WithHttpTransports(
				HttpTransportFunc(func(req *http.Request, next RoundTrip) (*http.Response, error) {
					applied = true
					req.Header.Set("X-Test", "1")
					return next(req)
				}),
			)(roundTripperFunc(func(req *http.Request) (*http.Response, error) {
				if req.Header.Get("X-Test") != "1" {
					return nil, errClient("transport not applied")
				}
				return &http.Response{StatusCode: http.StatusNoContent, Body: io.NopCloser(bytes.NewBuffer(nil)), Header: http.Header{}, Request: req}, nil
			}))
			_, err := rt.RoundTrip(mustRequest(http.MethodGet, "http://example.com", nil))
			if err != nil {
				return err
			}
			if !applied {
				return errClient("http transport was not invoked")
			}
			return nil
		}),
	)
}

func TestClientRequestAndDo(t0 *testing.T) {
	Then(t0, "客户端可构造请求并处理成功与失败响应",
		ExpectMust(func() error {
			c := &Client{Endpoint: "https://example.com/api"}
			req, err := c.newRequest(context.Background(), testRequest{ID: "1"}, courier.Metadata{"X-Trace": {"trace-1"}})
			if err != nil {
				return err
			}
			if req.URL.String() != "https://example.com/api/users/1" {
				return errClient("unexpected request url: " + req.URL.String())
			}
			if req.Header.Get("X-Trace") != "trace-1" {
				return errClient("unexpected request header")
			}
			return nil
		}),
		ExpectDo(func() error {
			c := &Client{Endpoint: "://bad"}
			_, err := c.newRequest(context.Background(), testRequest{ID: "1"})
			return err
		}, ErrorMatch(regexp.MustCompile(`InvalidEndpoint\{message="补全客户端 Endpoint 失败: .*missing protocol scheme",statusCode=400\}`))),
		ExpectMust(func() error {
			var seenHost string
			c := &Client{
				Endpoint: "https://example.com",
				HttpTransports: []HttpTransport{
					HttpTransportFunc(func(req *http.Request, next RoundTrip) (*http.Response, error) {
						seenHost = req.URL.Host
						return next(req)
					}),
				},
			}
			ctx := ContextWithHttpClient(context.Background(), &http.Client{
				Transport: roundTripperFunc(func(req *http.Request) (*http.Response, error) {
					return &http.Response{
						StatusCode: http.StatusOK,
						Header:     http.Header{"Content-Type": []string{"application/json"}},
						Body:       io.NopCloser(bytes.NewBufferString(`{"value":"ok"}`)),
						Request:    req,
					}, nil
				}),
			})
			meta, err := c.Do(ctx, testRequest{ID: "1"}).Into(&customContentType{})
			if err != nil {
				return err
			}
			if seenHost != "example.com" || meta.Get("Content-Type") != "application/json" {
				return errClient("unexpected transport/do result")
			}
			return nil
		}),
		ExpectMust(func() error {
			c := &Client{Endpoint: "https://example.com"}
			ctx := ContextWithHttpClient(context.Background(), &http.Client{
				Transport: roundTripperFunc(func(req *http.Request) (*http.Response, error) {
					return nil, context.Canceled
				}),
			})
			_, err := c.Do(ctx, testRequest{ID: "1"}).Into(nil)
			if err == nil || !regexp.MustCompile(`ClientClosedRequest\{message="请求已取消: .*context canceled",statusCode=499\}`).MatchString(err.Error()) {
				if err == nil {
					return errClient("expected client closed request error, got nil")
				}
				return errClient("expected client closed request error, got: " + err.Error())
			}
			return nil
		}),
		ExpectMust(func() error {
			c := &Client{Endpoint: "https://example.com"}
			ctx := ContextWithHttpClient(context.Background(), &http.Client{
				Transport: roundTripperFunc(func(req *http.Request) (*http.Response, error) {
					return nil, errors.New("boom")
				}),
			})
			_, err := c.Do(ctx, testRequest{ID: "1"}).Into(nil)
			if err == nil || !regexp.MustCompile(`RequestFailed\{message="发送请求失败: .*boom",statusCode=500\}`).MatchString(err.Error()) {
				return errClient("expected request failed error")
			}
			return nil
		}),
	)
}

func TestClientResultIntoBranches(t0 *testing.T) {
	req := mustRequest(http.MethodGet, "http://example.com", nil)

	Then(t0, "result.Into 覆盖常见分支",
		Expect((&result{Response: &http.Response{StatusCode: http.StatusCreated}}).StatusCode(), Equal(http.StatusCreated)),
		ExpectMust(func() error {
			r := &result{Response: &http.Response{Header: http.Header{"X-Test": []string{"1"}}}}
			if r.Meta().Get("X-Test") != "1" {
				return errClient("unexpected meta")
			}
			return nil
		}),
		ExpectMust(func() error {
			r := &result{
				Response: &http.Response{
					StatusCode: http.StatusOK,
					Header:     http.Header{"Content-Type": []string{"application/json; charset=utf-8"}},
					Body:       io.NopCloser(bytes.NewBufferString(`{"value":"ok"}`)),
					Request:    req,
				},
			}
			body := &customContentType{}
			_, err := r.Into(body)
			if err != nil {
				return err
			}
			if body.Value != "ok" {
				return errClient("unexpected decoded body")
			}
			return nil
		}),
		ExpectMust(func() error {
			r := &result{
				c: &Client{NewError: func() error { return &decodeError{} }},
				Response: &http.Response{
					StatusCode: http.StatusInternalServerError,
					Header:     http.Header{},
					Body:       io.NopCloser(bytes.NewBufferString(`failed`)),
					Request:    req,
				},
			}
			_, err := r.Into(&decodeError{})
			de, ok := err.(*decodeError)
			if !ok || de.statusCode != http.StatusInternalServerError || de.body != "failed" {
				return errClient("unexpected error body result")
			}
			return nil
		}),
		ExpectMust(func() error {
			r := &result{
				c: &Client{NewError: func() error { return &decodeError{} }},
				Response: &http.Response{
					StatusCode: http.StatusBadRequest,
					Header:     http.Header{},
					Body:       io.NopCloser(bytes.NewBufferString(`bad request`)),
					Request:    req,
				},
			}
			_, err := r.Into(nil)
			if err == nil || !regexp.MustCompile("bad request").MatchString(err.Error()) {
				return errClient("unexpected NewError result")
			}
			return nil
		}),
		ExpectMust(func() error {
			buf := bytes.NewBuffer(nil)
			r := &result{
				Response: &http.Response{
					StatusCode: http.StatusOK,
					Header:     http.Header{},
					Body:       io.NopCloser(bytes.NewBufferString(`stream`)),
					Request:    req,
				},
			}
			_, err := r.Into(buf)
			if err != nil || buf.String() != "stream" {
				return errClient("unexpected writer result")
			}
			return nil
		}),
		ExpectMust(func() error {
			var anything any
			r := &result{
				Response: &http.Response{
					StatusCode: http.StatusOK,
					Header:     http.Header{},
					Body:       io.NopCloser(bytes.NewBufferString(`ignored`)),
					Request:    req,
				},
			}
			_, err := r.Into(&anything)
			return err
		}),
		ExpectMust(func() error {
			r := &result{err: errors.New("boom")}
			_, err := r.Into(nil)
			if err == nil || err.Error() != "boom" {
				return errClient("unexpected direct error")
			}
			return nil
		}),
		ExpectMust(func() error {
			r := &result{
				Response: &http.Response{
					StatusCode: http.StatusOK,
					Header:     http.Header{},
					Body:       io.NopCloser(bytes.NewBufferString(`bad`)),
					Request:    req,
				},
			}
			errTarget := errors.New("target")
			_, err := r.Into(errTarget)
			if err == nil {
				return errClient("expected decode error")
			}
			return nil
		}),
	)
}

func TestHostAndDefaultClientHelpers(t0 *testing.T) {
	Then(t0, "host alias 与默认 client 辅助方法符合预期",
		Expect(HostAlias{}.IsZero(), Equal(true)),
		ExpectMust(func() error {
			hosts := Hosts{}
			hosts.AddHostAlias(HostAlias{IP: net.ParseIP("127.0.0.1"), Hostnames: []string{"example.com"}})
			var calledAddr string
			dial := hosts.WrapDialContext(func(ctx context.Context, network string, address string) (net.Conn, error) {
				calledAddr = address
				return nil, errors.New("stop")
			})
			_, _ = dial(context.Background(), "tcp", "example.com:80")
			if calledAddr != "127.0.0.1:80" {
				return errClient("unexpected resolved addr: " + calledAddr)
			}
			if selected := hosts.selectIP(func(yield func(string) bool) { yield("127.0.0.1") }, 1); selected != "127.0.0.1" {
				return errClient("unexpected selected ip: " + selected)
			}
			return nil
		}),
		ExpectMust(func() error {
			AddHostAlias(HostAlias{})
			SetDefaultTLSClientConfig(&tls.Config{ServerName: "example.com"})
			if reasonableRoundTripper.TLSClientConfig == nil || reasonableRoundTripper.TLSClientConfig.ServerName != "example.com" {
				return errClient("unexpected tls config")
			}
			return nil
		}),
		ExpectMust(func() error {
			ctx := ContextWithRoundTripperCreator(context.Background(), func() http.RoundTripper {
				return roundTripperFunc(func(req *http.Request) (*http.Response, error) { return nil, nil })
			})
			if _, ok := GetReasonableClientContext(ctx).Transport.(roundTripperFunc); !ok {
				return errClient("unexpected reasonable client transport")
			}
			if _, ok := GetShortConnClientContext(ctx).Transport.(roundTripperFunc); !ok {
				return errClient("unexpected short conn client transport")
			}
			return nil
		}),
		ExpectMust(func() error {
			rt := newRoundTripperWithoutKeepAlive()
			if x, ok := rt.(*http.Transport); !ok || !x.DisableKeepAlives {
				return errClient("unexpected short conn round tripper")
			}
			return nil
		}),
		ExpectMust(func() error {
			httpRT := &http.Transport{DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) { return nil, nil }}
			converted := convertTransportForH2c(httpRT)
			if _, ok := converted.(*http2.Transport); !ok {
				return errClient("expected h2c transport")
			}
			httpRT.DisableKeepAlives = true
			if convertTransportForH2c(httpRT) != httpRT {
				return errClient("expected same transport when keep alive disabled")
			}
			return nil
		}),
		ExpectMust(func() error {
			c := &Client{}
			if doc, ok := c.RuntimeDoc(); !ok || len(doc) != 0 {
				return errClient("unexpected runtime doc")
			}
			for _, name := range []string{"Endpoint", "NewError", "HttpTransports"} {
				if _, ok := c.RuntimeDoc(name); !ok {
					return errClient("missing runtime doc for " + name)
				}
			}
			if _, ok := c.RuntimeDoc("Unknown"); ok {
				return errClient("unexpected runtime doc hit")
			}
			if _, ok := runtimeDoc(struct{}{}, "", "Endpoint"); ok {
				return errClient("unexpected runtimeDoc helper hit")
			}
			return nil
		}),
		Expect(isOk(http.StatusOK), Equal(true)),
		Expect(isOk(http.StatusBadRequest), Equal(false)),
	)
}

func mustRequest(method string, rawURL string, body io.Reader) *http.Request {
	req, err := http.NewRequest(method, rawURL, body)
	if err != nil {
		panic(err)
	}
	return req
}

func errClient(msg string) error {
	return &clientErr{msg: msg}
}

type clientErr struct{ msg string }

func (e *clientErr) Error() string { return e.msg }
