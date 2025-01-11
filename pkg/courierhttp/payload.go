package courierhttp

import (
	"context"
	"fmt"
	"github.com/go-courier/logr"
	"github.com/octohelm/courier/internal/httprequest"
	"github.com/octohelm/courier/pkg/content"
	"github.com/octohelm/courier/pkg/courier"
	"github.com/octohelm/courier/pkg/statuserror"
	"io"
	"net/http"
	"net/textproto"
	"net/url"
	"reflect"
	"strconv"
)

type NoContent struct{}

type ContentTypeDescriber interface {
	ContentType() string
}

type StatusCodeDescriber interface {
	StatusCode() int
}

type CookiesDescriber interface {
	Cookies() []*http.Cookie
}

type RedirectDescriber interface {
	StatusCodeDescriber

	Location() *url.URL
}

type WithHeader interface {
	Header() http.Header
}

type FileHeader interface {
	io.ReadCloser
	Filename() string
	Header() http.Header
}

type RequestInfo = httprequest.Request

type ResponseSetting interface {
	SetStatusCode(statusCode int)
	SetLocation(location *url.URL)
	SetContentType(contentType string)
	SetMetadata(key string, values ...string)
	SetCookies(cookies []*http.Cookie)
}

type ResponseSettingFunc = func(s ResponseSetting)

func WithStatusCode(statusCode int) ResponseSettingFunc {
	return func(s ResponseSetting) {
		s.SetStatusCode(statusCode)
	}
}

func WithCookies(cookies ...*http.Cookie) ResponseSettingFunc {
	return func(s ResponseSetting) {
		s.SetCookies(cookies)
	}
}

func WithContentType(contentType string) ResponseSettingFunc {
	return func(s ResponseSetting) {
		s.SetContentType(contentType)
	}
}

func WithMetadata(key string, values ...string) ResponseSettingFunc {
	return func(s ResponseSetting) {
		s.SetMetadata(key, values...)
	}
}

func Wrap[T any](v T, opts ...ResponseSettingFunc) Response[T] {
	resp := &response[T]{
		v: v,
	}

	for i := range opts {
		opts[i](resp)
	}

	return resp
}

func WrapError(err error, opts ...ResponseSettingFunc) ErrorResponse {
	errResp := &errorResponse{}
	errResp.response.v = err
	for i := range opts {
		opts[i](&errResp.response)
	}
	return errResp
}

type ErrorResponse interface {
	Error() string
	Unwrap() error

	StatusCodeDescriber
	ContentTypeDescriber
	CookiesDescriber
	courier.MetadataCarrier
}

type Response[T any] interface {
	Underlying() T
	StatusCodeDescriber
	ContentTypeDescriber
	CookiesDescriber
	courier.MetadataCarrier
}

type ErrResponseWriter interface {
	WriteErr(ctx context.Context, rw http.ResponseWriter, req RequestInfo, err error)
}

type ResponseWriter interface {
	WriteResponse(ctx context.Context, rw http.ResponseWriter, req RequestInfo) error
}

type errorResponse struct {
	response[error]
}

func (e errorResponse) Error() string {
	return e.Underlying().Error()
}

func (e errorResponse) Unwrap() error {
	return e.Underlying()
}

type response[T any] struct {
	v           any
	meta        courier.Metadata
	cookies     []*http.Cookie
	location    *url.URL
	contentType string
	statusCode  int
}

func (r *response[T]) Underlying() T {
	return r.v.(T)
}

func (r *response[T]) Cookies() []*http.Cookie {
	return r.cookies
}

func (r *response[T]) SetStatusCode(statusCode int) {
	r.statusCode = statusCode
}

func (r *response[T]) SetContentType(contentType string) {
	r.contentType = contentType
}

func (r *response[T]) SetMetadata(key string, values ...string) {
	if r.meta == nil {
		r.meta = map[string][]string{}
	}
	r.meta[key] = values

	if r.contentType == "" || len(values) > 0 {
		if textproto.CanonicalMIMEHeaderKey(key) == "Content-Type" {
			r.contentType = values[0]
		}
	}
}

func (r *response[T]) SetCookies(cookies []*http.Cookie) {
	r.cookies = cookies
}

func (r *response[T]) SetLocation(location *url.URL) {
	r.location = location
}

func (r *response[T]) StatusCode() int {
	return r.statusCode
}

func (r *response[T]) ContentType() string {
	return r.contentType
}

func (r *response[T]) Meta() courier.Metadata {
	return r.meta
}

func (r *response[T]) WriteResponse(ctx context.Context, rw http.ResponseWriter, req RequestInfo) (finalErr error) {
	defer func() {
		if x, ok := r.v.(io.Closer); ok {
			// close again to avoid some leak issue
			_ = x.Close()
		}
		r.v = nil
		if finalErr != nil {
			logr.FromContext(ctx).Error(finalErr)
		}
	}()

	if respWriter, ok := r.v.(ResponseWriter); ok {
		return respWriter.WriteResponse(ctx, rw, req)
	}

	resp := r.v

	if err, ok := resp.(error); ok {
		opInfo, _ := OperationInfoFromContext(ctx)

		resp = statuserror.AsErrorResponse(err, opInfo.Server.UserAgent())
	}

	if statusCodeDescriber, ok := resp.(StatusCodeDescriber); ok {
		r.SetStatusCode(statusCodeDescriber.StatusCode())
	}

	if r.statusCode == 0 {
		if resp == nil {
			r.SetStatusCode(http.StatusNoContent)
		} else {
			if req.Method() == http.MethodPost {
				r.SetStatusCode(http.StatusCreated)
			} else {
				r.SetStatusCode(http.StatusOK)
			}
		}
	}

	if r.location == nil {
		if redirectDescriber, ok := resp.(RedirectDescriber); ok {
			r.SetStatusCode(redirectDescriber.StatusCode())
			r.SetLocation(redirectDescriber.Location())
		}
	}

	if r.meta != nil {
		header := rw.Header()
		for key, values := range r.meta {
			if len(values) == 1 {
				if v := values[0]; len(v) > 0 && v[0] == ',' {
					if hv := header.Get(key); hv != "" {
						header.Set(key, hv+v)
						continue
					}
				}
			}
			header[textproto.CanonicalMIMEHeaderKey(key)] = values
		}
	}

	if r.cookies != nil {
		for i := range r.cookies {
			cookie := r.cookies[i]
			if cookie != nil {
				http.SetCookie(rw, cookie)
			}
		}
	}

	if r.location != nil {
		http.Redirect(rw, req.Underlying(), r.location.String(), r.statusCode)
		return nil
	}

	if r.statusCode == http.StatusNoContent {
		rw.WriteHeader(r.statusCode)
		return nil
	}

	switch v := resp.(type) {
	case courier.Result:
		if r.contentType != "" {
			rw.Header().Set("Content-Type", r.contentType)
		}
		// forward result
		rw.WriteHeader(r.statusCode)
		if _, err := v.Into(rw); err != nil {
			return fmt.Errorf("forward failed: %w", err)
		}
	default:
		if resp == nil {
			// skip nil resp
			rw.WriteHeader(r.statusCode)
			return nil
		}

		t, err := content.New(reflect.TypeOf(resp), "", "marshal")
		if err != nil {
			return err
		}

		c, err := t.Prepare(ctx, resp)
		if err != nil {
			return err
		}
		defer c.Close()

		if ct := c.GetContentType(); ct != "" {
			rw.Header().Set("Content-Type", ct)
		}

		if r.contentType != "" {
			rw.Header().Set("Content-Type", r.contentType)
		}

		if i := c.GetContentLength(); i > -1 {
			rw.Header().Set("Content-Length", strconv.FormatInt(i, 10))
		}

		rw.WriteHeader(r.statusCode)

		if _, err := io.Copy(rw, c); err != nil {
			return err
		}
	}

	return nil
}
