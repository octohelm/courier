package courierhttp

import (
	"context"
	"io"
	"net/http"
	"net/textproto"
	"net/url"
	"reflect"

	"github.com/pkg/errors"

	"github.com/octohelm/courier/pkg/statuserror"

	"github.com/octohelm/courier/pkg/courier"
	transformer "github.com/octohelm/courier/pkg/transformer/core"
	typesutil "github.com/octohelm/x/types"
)

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

type Request interface {
	Context() context.Context

	ServiceName() string

	Method() string
	Path() string
	Header() http.Header
	Values(in string, name string) []string
	Body() io.ReadCloser

	Underlying() *http.Request
}

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
	WriteErr(ctx context.Context, rw http.ResponseWriter, req Request, statusErr *statuserror.StatusErr)
}

type contextErrResponseWriter struct{}

func ContextWithErrResponseWriter(ctx context.Context, errResponseWriter ErrResponseWriter) context.Context {
	return context.WithValue(ctx, contextErrResponseWriter{}, errResponseWriter)
}

func ErrResponseWriterFromContext(ctx context.Context) ErrResponseWriter {
	if writeErrResp, ok := ctx.Value(contextErrResponseWriter{}).(ErrResponseWriter); ok {
		return writeErrResp
	}
	return nil
}

func ErrResponseWriterFunc(fn func(ctx context.Context, rw http.ResponseWriter, req Request, statusErr *statuserror.StatusErr)) ErrResponseWriter {
	return &errResponseWriterFunc{fn: fn}
}

func (e *errResponseWriterFunc) WriteErr(ctx context.Context, rw http.ResponseWriter, req Request, statusErr *statuserror.StatusErr) {
	e.fn(ctx, rw, req, statusErr)
}

type errResponseWriterFunc struct {
	fn func(ctx context.Context, rw http.ResponseWriter, req Request, statusErr *statuserror.StatusErr)
}

type ResponseWriter interface {
	WriteResponse(ctx context.Context, rw http.ResponseWriter, req Request) error
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

func (r *response[T]) WriteResponse(ctx context.Context, rw http.ResponseWriter, req Request) error {
	defer func() {
		r.v = nil
	}()

	if respWriter, ok := r.v.(ResponseWriter); ok {
		return respWriter.WriteResponse(ctx, rw, req)
	}

	resp := r.v

	if err, ok := resp.(error); ok {
		statusErr, ok := statuserror.IsStatusErr(err)
		if !ok {
			if errors.Is(err, context.Canceled) {
				// https://httpstatuses.com/499
				statusErr = statuserror.Wrap(err, 499, "ContextCanceled")
			} else {
				statusErr = statuserror.Wrap(err, http.StatusInternalServerError, "InternalServerError")
			}
		}
		resp = statusErr.AppendSource(req.ServiceName())

		if errResponseWriter := ErrResponseWriterFromContext(ctx); errResponseWriter != nil {
			errResponseWriter.WriteErr(ctx, rw, req, statusErr)
			return nil
		}
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

	if r.contentType != "" {
		rw.Header().Set("Content-Type", r.contentType)
	}

	switch v := resp.(type) {
	case courier.Result:
		rw.WriteHeader(r.statusCode)
		if _, err := v.Into(rw); err != nil {
			return err
		}
	case io.Reader:
		rw.WriteHeader(r.statusCode)
		defer func() {
			if c, ok := v.(io.Closer); ok {
				_ = c.Close()
			}
		}()
		if _, err := io.Copy(rw, v); err != nil {
			return err
		}
	default:
		if resp == nil {
			// ship nil resp
			rw.WriteHeader(r.statusCode)
			return nil
		}
		tf, err := transformer.NewTransformer(ctx, typesutil.FromRType(reflect.TypeOf(resp)), transformer.Option{
			MIME: r.contentType,
		})
		if err != nil {
			return err
		}
		if err := tf.EncodeTo(transformer.ContextWithStatusCode(ctx, r.statusCode), rw, resp); err != nil {
			return err
		}
	}
	return nil
}
