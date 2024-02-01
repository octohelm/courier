package transport

import (
	"bytes"
	"context"
	"github.com/octohelm/courier/internal/pathpattern"
	"github.com/octohelm/courier/pkg/transformer"
	"io"
	"mime"
	"net/http"
	"net/textproto"
	"net/url"
	"reflect"
	"sort"
	"strings"
	"sync"

	"github.com/julienschmidt/httprouter"
	"github.com/octohelm/courier/pkg/courierhttp"
	"github.com/octohelm/courier/pkg/transformer/core"
	verrors "github.com/octohelm/courier/pkg/validator"
	contextx "github.com/octohelm/x/context"
	reflectx "github.com/octohelm/x/reflect"
	typesx "github.com/octohelm/x/types"
	"github.com/pkg/errors"
)

type OutgoingTransport interface {
	NewRequest(ctx context.Context, v any) (*http.Request, error)
}

func NewRequest(ctx context.Context, v any) (*http.Request, error) {
	ot, err := NewOutgoingTransport(ctx, v)
	if err != nil {
		return nil, err
	}

	return ot.NewRequest(ctx, v)
}

var outgoingTransports = sync.Map{}
var courierHttpPkgPath = reflect.TypeOf(courierhttp.Method{}).PkgPath()

func NewOutgoingTransport(ctx context.Context, r any) (OutgoingTransport, error) {
	typ := reflectx.Deref(reflect.TypeOf(r))

	if v, ok := outgoingTransports.Load(typ); ok {
		return v.(OutgoingTransport), nil
	}

	ot := &outgoingTransport{}

	ot.InParameters = map[string][]transformer.RequestParameter{}

	ot.Type = typ

	if methodDescriber, ok := r.(courierhttp.MethodDescriber); ok {
		ot.RouteMethod = methodDescriber.Method()
	}

	if pathDescriber, ok := r.(courierhttp.PathDescriber); ok {
		ot.RoutePath = pathDescriber.Path()
	}

	if ot.RoutePath == "" {
		if ot.Type.Kind() == reflect.Struct {
			for i := 0; i < ot.Type.NumField(); i++ {
				f := ot.Type.Field(i)

				if f.Anonymous && f.Type.PkgPath() == courierHttpPkgPath && strings.HasPrefix(f.Name, "Method") {
					if p, ok := f.Tag.Lookup("path"); ok {
						ot.RoutePath = pathpattern.NormalizePath(p)
					}
				}
			}
		}
	}

	err := core.EachRequestParameter(ctx, typesx.FromRType(ot.Type), func(rp *transformer.RequestParameter) {
		if rp.In == "" {
			return
		}
		ot.InParameters[rp.In] = append(ot.InParameters[rp.In], *rp)
	})

	outgoingTransports.Store(typ, ot)

	return ot, err
}

type outgoingTransport struct {
	RouteMethod  string
	RoutePath    string
	Type         reflect.Type
	InParameters map[string][]transformer.RequestParameter
}

func (t *outgoingTransport) Method() string {
	return t.RouteMethod
}

func (t *outgoingTransport) Path() string {
	return t.RoutePath
}

func (t *outgoingTransport) NewRequest(ctx context.Context, v any) (*http.Request, error) {
	typ := reflectx.Deref(reflect.TypeOf(v))

	if t.Type != typ {
		return nil, errors.Errorf("unmatched outgoingTransport, need %s but got %s", t.Type, typ)
	}

	method := t.Method()
	rawUrl := t.Path()

	errSet := verrors.NewErrorSet("")

	params := map[string]string{}
	query := url.Values{}
	header := http.Header{}
	cookies := url.Values{}
	var body io.ReadCloser

	rv, ok := v.(reflect.Value)
	if !ok {
		rv = reflect.ValueOf(v)
	}

	rv = reflectx.Indirect(rv)

	for in := range t.InParameters {
		parameters := t.InParameters[in]

		for i := range parameters {
			p := parameters[i]

			fieldValue := p.FieldValue(rv)

			if !fieldValue.IsValid() {
				continue
			}

			if p.In == "body" && body == nil {
				tryEncode := false

				switch fieldValue.Kind() {
				case reflect.Ptr:
					tryEncode = !fieldValue.IsNil()
				case reflect.Interface:
					tryEncode = !fieldValue.IsNil()
				default:
					tryEncode = true
				}

				if tryEncode {
					switch x := fieldValue.Interface().(type) {
					case io.ReadCloser:
						if header.Get("Content-Type") == "" {
							header.Set("Content-Type", p.Transformer.Names()[0])
						}
						body = x
					default:
						b := bytes.NewBuffer(nil)
						body = courierhttp.WrapReadCloser(b)
						err := p.Transformer.EncodeTo(ctx, core.WriterWithHeader(b, header), fieldValue)
						if err != nil {
							errSet.AddErr(err, p.Name)
						}
					}
				}
				continue
			}

			writers := core.NewStringBuilders()

			if err := core.Wrap(p.Transformer, &p.TransformerOption.CommonOption).EncodeTo(ctx, writers, fieldValue); err != nil {
				errSet.AddErr(err, p.Name)
				continue
			}

			values := writers.StringSlice()

			switch p.In {
			case "path":
				params[p.Name] = values[0]
			case "query":
				query[p.Name] = values
			case "header":
				header[textproto.CanonicalMIMEHeaderKey(p.Name)] = values
			case "cookie":
				cookies[p.Name] = values
			}
		}
	}

	req, err := http.NewRequestWithContext(courierhttp.ContextWithRouteDescriber(ctx, t), method, rawUrl, nil)
	if err != nil {
		return nil, err
	}

	if len(params) > 0 {
		req = req.WithContext(contextx.WithValue(req.Context(), httprouter.ParamsKey, params))
		req.URL.Path = core.StringifyPath(req.URL.Path, params)
	}

	if len(query) > 0 {
		if method == http.MethodGet && ShouldQueryInBodyForHttpGet(ctx) {
			header.Set("Content-Type", mime.FormatMediaType("application/x-www-form-urlencoded", map[string]string{
				"param": "value",
			}))
			body = io.NopCloser(bytes.NewBufferString(query.Encode()))
		} else {
			req.URL.RawQuery = query.Encode()
		}
	}

	req.Header = header

	if n := len(cookies); n > 0 {
		names := make([]string, n)
		i := 0
		for name := range cookies {
			names[i] = name
			i++
		}
		sort.Strings(names)

		for _, name := range names {
			values := cookies[name]
			for i := range values {
				req.AddCookie(&http.Cookie{
					Name:  name,
					Value: values[i],
				})
			}
		}
	}

	if body != nil {
		switch x := courierhttp.WrapReadCloser(body).(type) {
		case interface{ Len() int64 }:
			req.ContentLength = x.Len()
		}
		req.Body = body
	}

	return req, nil
}

type contextQueryInBody struct{}

func EnableQueryInBodyForHttpGet(ctx context.Context) context.Context {
	return contextx.WithValue(ctx, contextQueryInBody{}, true)
}

func ShouldQueryInBodyForHttpGet(ctx context.Context) bool {
	if v, ok := ctx.Value(contextQueryInBody{}).(bool); ok {
		return v
	}
	return false
}
