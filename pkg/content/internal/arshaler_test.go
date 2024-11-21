package internal_test

import (
	"bytes"
	"context"
	"net/http"
	"testing"

	"github.com/octohelm/courier/internal/httprequest"
	"github.com/octohelm/courier/internal/testingutil"
	"github.com/octohelm/courier/pkg/content/internal"
	testingx "github.com/octohelm/x/testing"

	_ "github.com/octohelm/courier/pkg/content/transformers"
)

func TestRequestArshaler(t *testing.T) {
	req, _ := http.NewRequest("GET", "http://0.0.0.0/users/{id}?filter=1&filter=2", bytes.NewBufferString("test"))
	req.Header.Set("Content-Type", "text/plain")
	req.Header.Set("Cookie", "cookie=xxx")

	op := &struct {
		ID          string   `name:"id" in:"path"`
		Filter      []string `name:"filter,omitzero" in:"query"`
		Limit       int      `name:"limit,omitzero" default:"10" in:"query"`
		ContentType string   `name:"Content-Type,omitzero" in:"header"`
		Cookie      string   `name:"cookie,omitzero" in:"cookie"`
		Data        string   `in:"body"`
	}{}

	err := internal.UnmarshalRequest(req.WithContext(httprequest.ContextWithPathValueGetter(context.Background(), httprequest.Params{
		"id": "1",
	})), op)
	testingx.Expect(t, err, testingx.BeNil[error]())

	testingx.Expect(t, op.ID, testingx.Be("1"))
	testingx.Expect(t, op.ContentType, testingx.Be("text/plain"))
	testingx.Expect(t, op.Limit, testingx.Be(10))
	testingx.Expect(t, op.Filter, testingx.Equal([]string{"1", "2"}))
	testingx.Expect(t, op.Data, testingx.Equal("test"))
	testingx.Expect(t, op.Cookie, testingx.Equal("xxx"))

	req2, err := internal.NewRequest(context.Background(), "GET", "/users/{id}", op)
	testingx.Expect(t, err, testingx.BeNil[error]())
	testingx.Expect(t, req2, testingutil.BeRequest(`
GET /users/1?filter=1&filter=2&limit=10 HTTP/1.1
Content-Length: 4
Content-Type: text/plain; charset=utf-8
Cookie: cookie=xxx

test
`))
}
