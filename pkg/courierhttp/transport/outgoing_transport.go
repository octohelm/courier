package transport

import (
	"context"
	"net/http"
	"reflect"
	"strings"

	"github.com/octohelm/courier/internal/pathpattern"
	"github.com/octohelm/courier/pkg/content"
	"github.com/octohelm/courier/pkg/courierhttp"
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

func NewOutgoingTransport(ctx context.Context, r any) (OutgoingTransport, error) {
	tpe := reflect.TypeOf(r)
	for tpe.Kind() == reflect.Ptr {
		tpe = tpe.Elem()
	}

	ot := &outgoingTransport{
		Type: tpe,
	}

	if methodDescriber, ok := r.(courierhttp.MethodDescriber); ok {
		ot.RouteMethod = methodDescriber.Method()
	}

	if pathDescriber, ok := r.(courierhttp.PathDescriber); ok {
		ot.RoutePath = pathDescriber.Path()
	}

	if ot.RoutePath == "" {
		if ot.Type.Kind() == reflect.Struct {
			ot.RoutePath = resolvePathInTag(ot.Type)
		}
	}

	return ot, nil
}

var courierHttpPkgPath = reflect.TypeFor[courierhttp.Method]().PkgPath()

func resolvePathInTag(tpe reflect.Type) string {
	for i := 0; i < tpe.NumField(); i++ {
		f := tpe.Field(i)

		if f.Anonymous {
			if f.Type.PkgPath() == courierHttpPkgPath && strings.HasPrefix(f.Name, "Method") {
				if p, ok := f.Tag.Lookup("path"); ok {
					return pathpattern.NormalizePath(p)
				}
			}

			// deep walk
			if f.Type.Kind() == reflect.Struct {
				return resolvePathInTag(f.Type)
			}
		}
	}
	return ""
}

type outgoingTransport struct {
	RouteMethod string
	RoutePath   string
	Type        reflect.Type
}

func (t *outgoingTransport) Method() string {
	return t.RouteMethod
}

func (t *outgoingTransport) Path() string {
	return t.RoutePath
}

func (t *outgoingTransport) NewRequest(ctx context.Context, v any) (*http.Request, error) {
	return content.NewRequest(ctx, t.Method(), t.Path(), v)
}
