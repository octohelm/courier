package client

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"reflect"
	"strings"
	"sync"

	"github.com/octohelm/courier/internal/httprequest"
	"github.com/octohelm/courier/pkg/content"
	"github.com/octohelm/courier/pkg/courier"
	"github.com/octohelm/courier/pkg/courierhttp/transport"
	"github.com/octohelm/courier/pkg/statuserror"
)

type HttpTransport = func(rt http.RoundTripper) http.RoundTripper

func WithHttpTransports(rts ...HttpTransport) func(rt http.RoundTripper) http.RoundTripper {
	return func(r http.RoundTripper) http.RoundTripper {
		for _, rt := range rts {
			r = rt(r)
		}
		return r
	}
}

type RoundTrip = func(request *http.Request) (*http.Response, error)

func HttpTransportFunc(round func(request *http.Request, next RoundTrip) (*http.Response, error)) HttpTransport {
	return func(rt http.RoundTripper) http.RoundTripper {
		return &httpTransportFunc{
			rt:    rt,
			round: round,
		}
	}
}

type httpTransportFunc struct {
	rt    http.RoundTripper
	round func(request *http.Request, next RoundTrip) (*http.Response, error)
}

func (h *httpTransportFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return h.round(request, h.rt.RoundTrip)
}

type Client struct {
	Endpoint string `flag:""`
	UseH2c   bool   `flag:",omitzero"`

	NewError       func() error
	HttpTransports []HttpTransport

	endpoint *url.URL
	parseErr error
	once     sync.Once
}

func (c *Client) completeEndpoint(u *url.URL) error {
	if u.Host == "" || u.Scheme == "" {
		c.once.Do(func() {
			endpoint, err := url.Parse(c.Endpoint)

			if endpoint != nil {
				if endpoint.Scheme == "h2c" {
					endpoint.Scheme = "http"
					c.UseH2c = true
				}
			}

			c.endpoint = endpoint
			c.parseErr = err
		})

		if c.parseErr != nil {
			return c.parseErr
		}

		u.Scheme = c.endpoint.Scheme
		u.Host = c.endpoint.Host
		u.Path = path.Clean(c.endpoint.Path + u.Path)
	}

	return nil
}

func (c *Client) Do(ctx context.Context, req any, metas ...courier.Metadata) courier.Result {
	httpReq, ok := req.(*http.Request)
	if !ok {
		r, err := c.newRequest(ctx, req, metas...)
		if err != nil {
			return &result{
				c:   c,
				err: statuserror.Wrap(err, http.StatusInternalServerError, "HttpRequestFailed"),
			}
		}
		httpReq = r
	}

	httpClient := HttpClientFromContext(ctx)
	if httpClient == nil {
		httpClient = GetReasonableClientContext(ctx)
	}

	if httpClient.Transport == nil {
		httpClient.Transport = reasonableRoundTripper
	}

	if c.UseH2c {
		httpClient.Transport = convertTransportForH2c(httpClient.Transport)
	}

	httpClient.Transport = WithHttpTransports(c.HttpTransports...)(httpClient.Transport)

	resp, err := httpClient.Do(httpReq)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			return &result{
				c:   c,
				err: statuserror.Wrap(err, 499, "ClientClosedRequest"),
			}
		}

		return &result{
			c:   c,
			err: statuserror.Wrap(err, http.StatusInternalServerError, "RequestFailed"),
		}
	}

	return &result{
		c:        c,
		Response: resp,
	}
}

func (c *Client) newRequest(ctx context.Context, r any, metas ...courier.Metadata) (*http.Request, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	req, err := transport.NewRequest(ctx, r)
	if err != nil {
		return nil, statuserror.Wrap(err, http.StatusBadRequest, "NewRequestFailed")
	}

	if err := c.completeEndpoint(req.URL); err != nil {
		return nil, statuserror.Wrap(err, http.StatusBadRequest, "InvalidEndpoint")
	}

	for k, vs := range courier.FromMetas(metas...) {
		for _, v := range vs {
			req.Header.Add(k, v)
		}
	}

	return req, nil
}

type result struct {
	*http.Response
	c   *Client
	err error
}

func (r *result) StatusCode() int {
	if r.Response != nil {
		return r.Response.StatusCode
	}
	return 0
}

func (r *result) Meta() courier.Metadata {
	if r.Response != nil {
		return courier.Metadata(r.Response.Header)
	}
	return courier.Metadata{}
}

func (r *result) Into(body any) (courier.Metadata, error) {
	defer func() {
		if r.Response != nil && r.Response.Body != nil {
			_ = r.Response.Body.Close()
		}
	}()

	if r.err != nil {
		return nil, r.err
	}

	meta := courier.Metadata(r.Response.Header)

	if !isOk(r.Response.StatusCode) {
		if r.c.NewError != nil {
			body = r.c.NewError()
		} else {
			body = &statuserror.Descriptor{
				Source: r.Response.Request.Host,
			}
		}
	}

	if body == nil {
		return meta, nil
	}

	switch x := body.(type) {
	case *any:
		return meta, nil
	case interface {
		error
		UnmarshalErrorResponse(statusCode int, respBody []byte) error
	}:
		if r.Response != nil && r.Response.Body != nil {
			data, err := io.ReadAll(r.Response.Body)
			if err != nil {
				return meta, err
			}
			if err := x.UnmarshalErrorResponse(r.Response.StatusCode, data); err != nil {
				return nil, err
			}
			return meta, x
		}
		if err := x.UnmarshalErrorResponse(r.Response.StatusCode, nil); err != nil {
			return nil, err
		}
		return meta, x
	case error:
		// to unmarshal status error
		if err := r.unmarshalInto(x); err != nil {
			return meta, err
		}
		return meta, x
	case io.Writer:
		if _, err := io.Copy(x, r.Response.Body); err != nil {
			return meta, statuserror.Wrap(err, http.StatusInternalServerError, "WriteFailed")
		}
	default:
		if err := r.unmarshalInto(body); err != nil {
			return meta, err
		}
	}

	return meta, nil
}

func (r *result) unmarshalInto(body any) error {
	rv := reflect.ValueOf(body)

	mediaType := strings.Split(r.Response.Header.Get("Content-Type"), ";")[0]
	if v, ok := body.(interface{ ContentType() string }); ok {
		mediaType = v.ContentType()
	}

	tf, err := content.New(rv.Type(), mediaType, "unmarshal")
	if err != nil {
		return err
	}

	if err := tf.ReadAs(context.Background(), httprequest.WithHeader(r.Response.Body, r.Response.Header), rv); err != nil {
		return statuserror.Wrap(fmt.Errorf("unmarshal to %T failed: %w", body, err), http.StatusInternalServerError, "ResponseDecodeFailed")
	}

	return nil
}

func isOk(code int) bool {
	return code >= http.StatusOK && code < http.StatusMultipleChoices
}
