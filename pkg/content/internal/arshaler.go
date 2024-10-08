package internal

import (
	"context"
	"errors"
	"net/http"
	"reflect"

	"github.com/octohelm/courier/internal/httprequest"

	"github.com/octohelm/courier/internal/pathpattern"
)

func NewRequest(ctx context.Context, method string, path string, v any) (*http.Request, error) {
	rv := reflect.ValueOf(v)
	for rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}

	ra := &Request{}
	ra.Value = rv

	return ra.MarshalRequest(ctx, method, pathpattern.Parse(path))
}

func UnmarshalRequest(req *http.Request, out any) error {
	rv := reflect.ValueOf(out)
	if rv.Kind() != reflect.Ptr {
		return errors.New("unmarshal request target must be ptr value")
	}

	ra := &Request{}
	ra.Value = rv.Elem()

	return ra.UnmarshalRequest(req)
}

func UnmarshalRequestInfo(ireq httprequest.Request, out any) error {
	rv := reflect.ValueOf(out)
	if rv.Kind() != reflect.Ptr {
		return errors.New("unmarshal request target must be ptr value")
	}

	ra := &Request{}
	ra.Value = rv.Elem()

	return ra.UnmarshalRequestInfo(ireq)
}
