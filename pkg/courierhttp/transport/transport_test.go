package transport_test

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	. "github.com/octohelm/x/testing/v2"

	"github.com/octohelm/courier/internal/httprequest"
	"github.com/octohelm/courier/internal/testingutil"
	"github.com/octohelm/courier/pkg/courierhttp"
	"github.com/octohelm/courier/pkg/courierhttp/handler"
	"github.com/octohelm/courier/pkg/courierhttp/transport"
)

func TestRequestTransformer(t *testing.T) {
	type Headers struct {
		HInt    int    `in:"header"`
		HString string `in:"header"`
		HBool   bool   `in:"header"`
	}

	type Queries struct {
		QInt            int                   `name:"int" in:"query"`
		QEmptyInt       int                   `name:"emptyInt,omitzero" in:"query"`
		QString         string                `name:"string" in:"query"`
		QSlice          []string              `name:"slice" in:"query"`
		QBytes          []byte                `name:"bytes,omitzero" in:"query"`
		StartedAt       *testingutil.Datetime `name:"startedAt,omitzero" in:"query"`
		QBytesOmitEmpty []byte                `name:"bytesOmit,omitzero" in:"query"`
	}

	type Cookies struct {
		CString string   `name:"a" in:"cookie"`
		CSlice  []string `name:"slice" in:"cookie"`
	}

	type Data struct {
		A string `json:",omitzero" xml:",omitzero"`
		B string `json:",omitzero" xml:",omitzero"`
		C string `json:",omitzero" xml:",omitzero"`
	}

	t.Run("full parameter mapping", func(t *testing.T) {
		type Request struct {
			courierhttp.MethodGet `path:"/:id"`
			ID                    string `name:"id" in:"path"`
			Headers
			Queries
			Cookies

			Data `in:"body"`
		}

		req := Request{
			ID: "1",
			Headers: Headers{
				HInt:    1,
				HString: "string",
				HBool:   true,
			},
			Queries: Queries{
				QInt:    1,
				QString: "string",
				QSlice:  []string{"1", "2"},
				QBytes:  []byte("bytes"),
			},
			Cookies: Cookies{
				CString: "xxx",
				CSlice:  []string{"1", "2"},
			},
		}

		Then(t, "请求可在 outgoing 与 incoming transport 之间往返", ExpectMust(func() error {
			r, err := transport.NewRequest(context.Background(), req)
			if err != nil {
				return err
			}
			if err := testingutil.BeRequest(`
GET /1?bytes=bytes&int=1&slice=1&slice=2&string=string HTTP/1.1
Content-Type: application/json; charset=utf-8
Cookie: a=xxx; slice=1; slice=2
Hbool: true
Hint: 1
Hstring: string

{}
`)(r); err != nil {
				return err
			}

			req2 := &Request{}

			incomeTransport, err := transport.NewIncomingTransport(context.Background(), req2)
			if err != nil {
				return err
			}

			request := r.WithContext(httprequest.ContextWithPathValueGetter(r.Context(), handler.Params{"id": "1"}))

			if err := incomeTransport.UnmarshalOperator(context.Background(), httprequest.From(request), req2); err != nil {
				return err
			}
			return nil
		}))
	})

	t.Run("header fallback from query prefix", func(t *testing.T) {
		req := &struct {
			courierhttp.MethodGet `path:"/"`
			Headers
		}{}

		Then(t, "x-param-header- 前缀的 query 参数会映射为 header", ExpectMust(func() error {
			it, err := transport.NewIncomingTransport(context.Background(), req)
			if err != nil {
				return err
			}

			httpRequest, err := http.NewRequest("GET", "/?x-param-header-Hint=1&x-param-header-Hbool=true&x-param-header-Hstring=string", nil)
			if err != nil {
				return err
			}

			if err := it.UnmarshalOperator(context.Background(), httprequest.From(httpRequest), req); err != nil {
				return err
			}
			expected := Headers{HInt: 1, HString: "string", HBool: true}
			if req.Headers != expected {
				return fmt.Errorf("unexpected headers: %#v", req.Headers)
			}
			return nil
		}))
	})
}

func TestRequestTransformerAppliesDefaultValue(t *testing.T) {
	t.Run("uses provided query value when in range", func(t *testing.T) {
		req := &struct {
			courierhttp.MethodGet `path:"/"`
			Limit                 int64 `name:"limit,omitzero" validate:"@int[-1,50] = 10" in:"query"`
		}{}

		Then(t, "显式传值时保留请求值", ExpectMust(func() error {
			it, err := transport.NewIncomingTransport(context.Background(), req)
			if err != nil {
				return err
			}
			httpRequest, err := http.NewRequest("GET", "/?limit=20", nil)
			if err != nil {
				return err
			}
			if err := it.UnmarshalOperator(context.Background(), httprequest.From(httpRequest), req); err != nil {
				return err
			}
			if req.Limit != 20 {
				return fmt.Errorf("unexpected limit: %d", req.Limit)
			}
			return nil
		}))
	})

	t.Run("fills default value when query missing", func(t *testing.T) {
		req := &struct {
			courierhttp.MethodGet `path:"/"`
			Limit                 int64 `name:"limit,omitzero" validate:"@int[-1,50] = 10" in:"query"`
		}{}

		Then(t, "缺省值会在请求中缺失时补齐", ExpectMust(func() error {
			it, err := transport.NewIncomingTransport(context.Background(), req)
			if err != nil {
				return err
			}
			httpRequest, err := http.NewRequest("GET", "/", nil)
			if err != nil {
				return err
			}
			if err := it.UnmarshalOperator(context.Background(), httprequest.From(httpRequest), req); err != nil {
				return err
			}
			if req.Limit != 10 {
				return fmt.Errorf("unexpected default limit: %d", req.Limit)
			}
			return nil
		}))
	})

	t.Run("rejects out-of-range value", func(t *testing.T) {
		req := &struct {
			courierhttp.MethodGet `path:"/"`
			Limit                 int64 `name:"limit,omitzero" validate:"@int[-1,50] = 10" in:"query"`
		}{}

		Then(t, "超出范围的值会返回错误",
			ExpectMust(func() error {
				it, err := transport.NewIncomingTransport(context.Background(), req)
				if err != nil {
					return err
				}
				httpRequest, err := http.NewRequest("GET", "/?limit=200", nil)
				if err != nil {
					return err
				}
				if err := it.UnmarshalOperator(context.Background(), httprequest.From(httpRequest), req); err == nil {
					return fmt.Errorf("expected out-of-range error")
				}
				return nil
			}),
		)
	})
}
