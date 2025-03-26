package transport_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/octohelm/x/testing/bdd"

	"github.com/octohelm/courier/internal/httprequest"
	"github.com/octohelm/courier/internal/testingutil"
	"github.com/octohelm/courier/pkg/courierhttp"
	"github.com/octohelm/courier/pkg/courierhttp/handler"
	"github.com/octohelm/courier/pkg/courierhttp/transport"
	testingx "github.com/octohelm/x/testing"
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

	t.Run("full in parameters", func(t *testing.T) {
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

		r, err := transport.NewRequest(context.Background(), req)
		testingx.Expect(t, err, testingx.BeNil[error]())
		testingx.Expect(t, r, testingutil.BeRequest(`
GET /1?bytes=bytes&int=1&slice=1&slice=2&string=string HTTP/1.1
Content-Type: application/json; charset=utf-8
Cookie: a=xxx; slice=1; slice=2
Hbool: true
Hint: 1
Hstring: string

{}
`))

		req2 := &Request{}

		incomeTransport, err := transport.NewIncomingTransport(context.Background(), req2)
		testingx.Expect(t, err, testingx.BeNil[error]())

		request := r.WithContext(httprequest.ContextWithPathValueGetter(r.Context(), handler.Params{"id": "1"}))

		err = incomeTransport.UnmarshalOperator(context.Background(), httprequest.From(request), req2)
		testingx.Expect(t, err, testingx.BeNil[error]())
	})

	t.Run("Should unmarshal header values from query values with prefix `x-param-header-`", func(t *testing.T) {
		req := &struct {
			courierhttp.MethodGet `path:"/"`
			Headers
		}{}

		it, err := transport.NewIncomingTransport(context.Background(), req)
		testingx.Expect(t, err, testingx.Be[error](nil))

		httpRequest, err := http.NewRequest("GET", "/?x-param-header-Hint=1&x-param-header-Hbool=true&x-param-header-Hstring=string", nil)
		testingx.Expect(t, err, testingx.Be[error](nil))

		err = it.UnmarshalOperator(context.Background(), httprequest.From(httpRequest), req)
		testingx.Expect(t, err, testingx.Be[error](nil))
		testingx.Expect(t, req.Headers, testingx.Equal(Headers{
			HInt:    1,
			HString: "string",
			HBool:   true,
		}))
	})
}

func TestRequestTransformerWithDefaultValue(t *testing.T) {
	t.Run("declare a operator", bdd.GivenT(func(b bdd.T) {
		req := &struct {
			courierhttp.MethodGet `path:"/"`
			Limit                 int64 `name:"limit,omitzero" validate:"@int[-1,50] = 10" in:"query"`
		}{}

		it := bdd.Must(transport.NewIncomingTransport(b.Context(), req))

		b.When("handle request with limit in range", func(b bdd.T) {
			httpRequest := bdd.Must(http.NewRequest("GET", "/?limit=20", nil))

			b.Then("parse parameters successful",
				bdd.Nil(it.UnmarshalOperator(b.Context(), httprequest.From(httpRequest), req)),
			)

			b.Then("value should set",
				bdd.Equal(20, req.Limit),
			)
		})

		b.When("handle request with empty limit", func(b bdd.T) {
			httpRequest := bdd.Must(http.NewRequest("GET", "/", nil))

			b.Then("parse parameters successful",
				bdd.Nil(it.UnmarshalOperator(b.Context(), httprequest.From(httpRequest), req)),
			)

			b.Then("default value should set",
				bdd.Equal(10, req.Limit),
			)
		})

		b.When("handle request with limit which out of range", func(b bdd.T) {
			httpRequest := bdd.Must(http.NewRequest("GET", "/?limit=200", nil))

			b.Then("parse parameters failed",
				bdd.HasError(it.UnmarshalOperator(b.Context(), httprequest.From(httpRequest), req)),
			)
		})
	}))
}
