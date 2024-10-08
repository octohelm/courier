package httprequest

import (
	"context"
	"io"
	"net/http"
)

type Request interface {
	Context() context.Context
	Method() string
	Path() string
	Header() http.Header
	Values(in string, name string) []string
	Body() io.ReadCloser
	Underlying() *http.Request
}
