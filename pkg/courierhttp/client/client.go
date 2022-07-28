package client

import (
	"context"
	"io"
	"net/http"
	"net/textproto"
	"net/url"
	"reflect"
	"strings"

	"github.com/julienschmidt/httprouter"
	"github.com/octohelm/courier/pkg/courier"
	"github.com/octohelm/courier/pkg/courierhttp/transport"
	"github.com/octohelm/courier/pkg/statuserror"
	transformer "github.com/octohelm/courier/pkg/transformer/core"
	typesutil "github.com/octohelm/x/types"
	"github.com/pkg/errors"
)

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

type HttpTransport func(rt http.RoundTripper) http.RoundTripper

type Client struct {
	Endpoint       string `env:""`
	HttpTransports []HttpTransport
}

func (c *Client) Do(ctx context.Context, req any, metas ...courier.Metadata) courier.Result {
	httpReq, ok := req.(*http.Request)
	if !ok {
		r, err := c.newRequest(ctx, req, metas...)
		if err != nil {
			return &result{
				err: statuserror.Wrap(err, http.StatusInternalServerError, "RequestFailed"),
			}
		}
		httpReq = r
	}

	httpClient := HttpClientFromContext(ctx)
	if httpClient == nil {
		httpClient = GetShortConnClientContext(ctx, c.HttpTransports...)
	}

	resp, err := httpClient.Do(httpReq)
	if err != nil {
		if errors.Unwrap(err) == context.Canceled {
			return &result{
				err: statuserror.Wrap(err, 499, "ClientClosedRequest"),
			}
		}

		return &result{
			err: statuserror.Wrap(err, http.StatusInternalServerError, "RequestFailed"),
		}
	}

	return &result{
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

	if u := req.URL.String(); !strings.HasPrefix(u, c.Endpoint) {
		uu, _ := url.Parse(c.Endpoint)
		req.URL.Scheme = uu.Scheme
		req.URL.Host = uu.Host
		req.URL.Path = httprouter.CleanPath(uu.Path + req.URL.Path)
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
		body = &statuserror.StatusErr{
			Code:    r.Response.StatusCode,
			Msg:     r.Response.Status,
			Sources: []string{r.Response.Request.Host},
		}
	}

	if body == nil {
		return meta, nil
	}

	switch x := body.(type) {
	case error:
		// to unmarshal status error
		if err := r.decode(x, meta); err != nil {
			return meta, err
		}
		return meta, x
	case io.Writer:
		if _, err := io.Copy(x, r.Response.Body); err != nil {
			return meta, statuserror.Wrap(err, http.StatusInternalServerError, "WriteFailed")
		}
	default:
		if err := r.decode(body, meta); err != nil {
			return meta, err
		}
	}

	return meta, nil
}

func (r *result) decode(body any, meta courier.Metadata) error {
	rv := reflect.ValueOf(body)

	tf, err := transformer.NewTransformer(context.Background(), typesutil.FromRType(rv.Type()), transformer.Option{})
	if err != nil {
		return statuserror.Wrap(err, http.StatusInternalServerError, "TransformerCreateFailed")
	}

	if err := tf.DecodeFrom(context.Background(), r.Response.Body, rv, textproto.MIMEHeader(r.Response.Header)); err != nil {
		return statuserror.Wrap(err, http.StatusInternalServerError, "DecodeFailed", errors.Wrapf(err, "decode failed to %v", body).Error())
	}

	return nil
}

func isOk(code int) bool {
	return code >= http.StatusOK && code < http.StatusMultipleChoices
}
