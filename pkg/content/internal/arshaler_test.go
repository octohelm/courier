package internal_test

import (
	"bytes"
	"context"
	"net/http"
	"testing"

	. "github.com/octohelm/x/testing/v2"

	"github.com/octohelm/courier/internal/httprequest"
	"github.com/octohelm/courier/internal/testingutil"
	"github.com/octohelm/courier/pkg/content/internal"
)

import (
	_ "github.com/octohelm/courier/pkg/content/transformers"
)

func TestRequestArshalerRoundTrip(t *testing.T) {
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

	Then(t, "请求字段可在 HTTP 请求和 operator 之间双向映射", ExpectMust(func() error {
		err := internal.UnmarshalRequest(req.WithContext(httprequest.ContextWithPathValueGetter(context.Background(), httprequest.Params{
			"id": "1",
		})), op)
		if err != nil {
			return err
		}

		if op.ID != "1" || op.ContentType != "text/plain" || op.Limit != 10 || op.Data != "test" || op.Cookie != "xxx" {
			return errArshaler("unexpected request values")
		}
		if len(op.Filter) != 2 || op.Filter[0] != "1" || op.Filter[1] != "2" {
			return errArshaler("unexpected filter values")
		}

		req2, err := internal.NewRequest(context.Background(), "GET", "/users/{id}", op)
		if err != nil {
			return err
		}
		return testingutil.BeRequest(`
GET /users/1?filter=1&filter=2&limit=10 HTTP/1.1
Content-Length: 4
Content-Type: text/plain; charset=utf-8
Cookie: cookie=xxx

test
`)(req2)
	}))
}

func TestUnmarshalRequestInfoWrapper(t *testing.T) {
	t.Run("requires pointer target", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "http://0.0.0.0/users/1", nil)

		Then(t, "UnmarshalRequestInfo 对非指针目标返回错误", ExpectMust(func() error {
			err := internal.UnmarshalRequestInfo(httprequest.From(req), struct{}{})
			if err == nil || err.Error() != "unmarshal request target must be ptr value" {
				return errArshaler("unexpected non-pointer error")
			}
			return nil
		}))
	})

	t.Run("delegates to request info decoder", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "http://0.0.0.0/users/1?filter=1&filter=2", bytes.NewBufferString("test"))
		req.Header.Set("Content-Type", "text/plain")

		op := &struct {
			ID     string   `name:"id" in:"path"`
			Filter []string `name:"filter,omitzero" in:"query"`
			Data   string   `in:"body"`
		}{}

		Then(t, "UnmarshalRequestInfo 可直接完成 path、query、body 解码", ExpectMust(func() error {
			err := internal.UnmarshalRequestInfo(httprequest.From(req.WithContext(httprequest.ContextWithPathValueGetter(context.Background(), httprequest.Params{
				"id": "1",
			}))), op)
			if err != nil {
				return err
			}
			if op.ID != "1" || op.Data != "test" {
				return errArshaler("unexpected request info values")
			}
			if len(op.Filter) != 2 || op.Filter[0] != "1" || op.Filter[1] != "2" {
				return errArshaler("unexpected request info filter values")
			}
			return nil
		}))
	})
}

func errArshaler(msg string) error {
	return &arshalerErr{msg: msg}
}

type arshalerErr struct{ msg string }

func (e *arshalerErr) Error() string { return e.msg }
