package content

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"testing"

	. "github.com/octohelm/x/testing/v2"

	"github.com/octohelm/courier/internal/httprequest"
)

type textValue string

func (v textValue) MarshalText() ([]byte, error) {
	return []byte(v), nil
}

func (v *textValue) UnmarshalText(data []byte) error {
	*v = textValue(data)
	return nil
}

type jsonValue struct {
	Name string `json:"name"`
}

type closeableReader struct {
	*bytes.Reader
}

func (r *closeableReader) Close() error {
	return nil
}

func TestNew(t0 *testing.T) {
	Then(t0, "按类型和媒体类型选择转换器",
		ExpectMustValue(func() (string, error) {
			x, err := New(reflect.TypeFor[string](), "", "marshal")
			if err != nil {
				return "", err
			}
			return x.MediaType(), nil
		}, Equal("text/plain")),
		ExpectMustValue(func() (string, error) {
			x, err := New(reflect.TypeFor[textValue](), "", "marshal")
			if err != nil {
				return "", err
			}
			return x.MediaType(), nil
		}, Equal("text/plain")),
		ExpectMustValue(func() (string, error) {
			x, err := New(reflect.TypeFor[jsonValue](), "application/problem+json", "marshal")
			if err != nil {
				return "", err
			}
			return x.MediaType(), nil
		}, Equal("application/json")),
		ExpectMustValue(func() (string, error) {
			x, err := New(reflect.TypeFor[*closeableReader](), "", "marshal")
			if err != nil {
				return "", err
			}
			return x.MediaType(), nil
		}, Equal("application/octet-stream")),
	)
}

func TestRequestRoundTrip(t0 *testing.T) {
	type request struct {
		ID          string   `name:"id" in:"path"`
		Filter      []string `name:"filter,omitzero" in:"query"`
		Limit       int      `name:"limit,omitzero" default:"10" in:"query"`
		ContentType string   `name:"Content-Type,omitzero" in:"header"`
		Cookie      string   `name:"cookie,omitzero" in:"cookie"`
		Data        string   `in:"body"`
	}

	origin := request{
		ID:          "1",
		Filter:      []string{"1", "2"},
		Limit:       10,
		ContentType: "text/plain",
		Cookie:      "xxx",
		Data:        "test",
	}

	Then(t0, "从结构体生成请求并反序列化回来",
		ExpectMustValue(func() (*http.Request, error) {
			return NewRequest(context.Background(), http.MethodGet, "/users/{id}", origin)
		}, Be(func(req *http.Request) error {
			body, err := io.ReadAll(req.Body)
			if err != nil {
				return err
			}

			if req.Method != http.MethodGet {
				return fmt.Errorf("unexpected method %s", req.Method)
			}
			if req.URL.Path != "/users/1" {
				return fmt.Errorf("unexpected path %s", req.URL.Path)
			}
			if req.URL.RawQuery != "filter=1&filter=2&limit=10" {
				return fmt.Errorf("unexpected query %s", req.URL.RawQuery)
			}
			if req.Header.Get("Content-Type") != "text/plain; charset=utf-8" {
				return fmt.Errorf("unexpected content-type %s", req.Header.Get("Content-Type"))
			}
			if req.Header.Get("Cookie") != "cookie=xxx" {
				return fmt.Errorf("unexpected cookie header %s", req.Header.Get("Cookie"))
			}
			if string(body) != "test" {
				return fmt.Errorf("unexpected body %q", string(body))
			}
			return nil
		})),
		ExpectMust(func() error {
			req, err := NewRequest(context.Background(), http.MethodGet, "/users/{id}", origin)
			if err != nil {
				return err
			}

			target := &request{}
			err = UnmarshalRequest(req.WithContext(httprequest.ContextWithPathValueGetter(req.Context(), httprequest.Params{
				"id": "1",
			})), target)
			if err != nil {
				return err
			}

			if target.ID != origin.ID || !reflect.DeepEqual(target.Filter, origin.Filter) || target.Limit != origin.Limit || target.Cookie != origin.Cookie || target.Data != origin.Data {
				return fmt.Errorf("unexpected request %#v", target)
			}
			if target.ContentType != "text/plain; charset=utf-8" {
				return fmt.Errorf("unexpected content-type %q", target.ContentType)
			}

			return nil
		}),
		ExpectMust(func() error {
			req, err := NewRequest(context.Background(), http.MethodGet, "/users/{id}", origin)
			if err != nil {
				return err
			}
			reqInfo := httprequest.From(req.WithContext(httprequest.ContextWithPathValueGetter(req.Context(), httprequest.Params{
				"id": "1",
			})))

			target := &request{}
			if err := UnmarshalRequestInfo(reqInfo, target); err != nil {
				return err
			}
			if target.ID != "1" || target.Cookie != "xxx" {
				return fmt.Errorf("unexpected request info %#v", target)
			}
			return nil
		}),
	)
}
