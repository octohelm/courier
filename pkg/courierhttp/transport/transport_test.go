package transport_test

import (
	"context"
	"reflect"
	"testing"

	"github.com/octohelm/courier/internal/testingutil"
	"github.com/octohelm/courier/pkg/courierhttp"
	"github.com/octohelm/courier/pkg/courierhttp/handler"
	"github.com/octohelm/courier/pkg/courierhttp/transport"
	reflectx "github.com/octohelm/x/reflect"
)

func TestRequestTransformer(t *testing.T) {
	type Headers struct {
		HInt    int    `in:"header"`
		HString string `in:"header"`
		HBool   bool   `in:"header"`
	}

	type Queries struct {
		QInt            int                   `name:"int" in:"query"`
		QEmptyInt       int                   `name:"emptyInt,omitempty" in:"query"`
		QString         string                `name:"string" in:"query"`
		QSlice          []string              `name:"slice" in:"query"`
		QBytes          []byte                `name:"bytes,omitempty" in:"query"`
		StartedAt       *testingutil.Datetime `name:"startedAt,omitempty" in:"query"`
		QBytesOmitEmpty []byte                `name:"bytesOmit,omitempty" in:"query"`
	}

	type Cookies struct {
		CString string   `name:"a" in:"cookie"`
		CSlice  []string `name:"slice" in:"cookie"`
	}

	type Data struct {
		A string `json:",omitempty" xml:",omitempty"`
		B string `json:",omitempty" xml:",omitempty"`
		C string `json:",omitempty" xml:",omitempty"`
	}

	type FormDataMultipart struct {
		Bytes []byte `name:"bytes"`
		A     []int  `name:"a"`
		C     uint   `name:"c" `
		Data  Data   `name:"data"`
	}

	cases := []struct {
		name   string
		path   string
		expect string
		req    interface{}
	}{
		{
			"full InParameters",
			"/:id",
			`GET /1?bytes=Ynl0ZXM%3D&int=1&slice=1&slice=2&string=string HTTP/1.1
Content-Type: application/json; charset=utf-8
Cookie: a=xxx; slice=1; slice=2
Hbool: true
Hint: 1
Hstring: string

{}
`,
			&struct {
				courierhttp.MethodGet `path:"/:id"`
				Headers
				Queries
				Cookies
				Data `in:"body"`
				ID   string `name:"id" in:"path"`
			}{
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
			},
		},
		{
			"url-encoded",
			"/",
			`GET / HTTP/1.1
Content-Type: application/x-www-form-urlencoded; param=value

int=1&slice=1&slice=2&string=string`,
			&struct {
				courierhttp.MethodGet `path:"/"`
				Queries               `in:"body" mime:"urlencoded"`
			}{
				Queries: Queries{
					QInt:    1,
					QString: "string",
					QSlice:  []string{"1", "2"},
				},
			},
		},
		{
			"xml",
			"/",
			`GET / HTTP/1.1
Content-Type: application/xml; charset=utf-8

<Data><A>1</A></Data>`,
			&struct {
				courierhttp.MethodGet `path:"/"`
				Data                  `in:"body" mime:"xml"`
			}{
				Data: Data{
					A: "1",
				},
			},
		},
		{
			"form-data/multipart",
			"/",
			`GET / HTTP/1.1
Content-Type: multipart/form-data; boundary=5eaf397248958ac38281d1c034e1ad0d4a5f7d986d4c53ac32e8399cbcda

--5eaf397248958ac38281d1c034e1ad0d4a5f7d986d4c53ac32e8399cbcda
Content-Disposition: form-data; name="bytes"
Content-Type: text/plain; charset=utf-8

Ynl0ZXM=
--5eaf397248958ac38281d1c034e1ad0d4a5f7d986d4c53ac32e8399cbcda
Content-Disposition: form-data; name="a"
Content-Type: text/plain; charset=utf-8

-1
--5eaf397248958ac38281d1c034e1ad0d4a5f7d986d4c53ac32e8399cbcda
Content-Disposition: form-data; name="a"
Content-Type: text/plain; charset=utf-8

1
--5eaf397248958ac38281d1c034e1ad0d4a5f7d986d4c53ac32e8399cbcda
Content-Disposition: form-data; name="c"
Content-Type: text/plain; charset=utf-8

1
--5eaf397248958ac38281d1c034e1ad0d4a5f7d986d4c53ac32e8399cbcda
Content-Disposition: form-data; name="data"
Content-Type: application/json; charset=utf-8

{"A":"1"}

--5eaf397248958ac38281d1c034e1ad0d4a5f7d986d4c53ac32e8399cbcda--
`,
			&struct {
				courierhttp.MethodGet `path:"/"`
				FormDataMultipart     `in:"body" mime:"multipart" boundary:"boundary1"`
			}{
				FormDataMultipart: FormDataMultipart{
					A:     []int{-1, 1},
					C:     1,
					Bytes: []byte("bytes"),
					Data: Data{
						A: "1",
					},
				},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			for i := 0; i < 5; i++ {
				t.Run("New outgoing request", func(t *testing.T) {
					req, err := transport.NewRequest(context.Background(), c.req)
					testingutil.Expect(t, err, testingutil.Be[error](nil))

					testingutil.RequestEqual(t, req, c.expect)

					t.Run("Unmarshal incoming request", func(t *testing.T) {
						rv := reflectx.New(reflect.PtrTo(reflectx.Deref(reflect.TypeOf(c.req))))

						it, err := transport.NewIncomingTransport(context.Background(), rv.Interface())
						testingutil.Expect(t, err, testingutil.Be[error](nil))

						req = req.WithContext(handler.ContextWithParamGetter(req.Context(), handler.Params{"id": "1"}))

						err = it.UnmarshalOperator(context.Background(), transport.FromHttpRequest(req, ""), rv)
						testingutil.Expect(t, err, testingutil.Be[error](nil))
						testingutil.Expect(t, reflectx.Indirect(rv).Interface(), testingutil.Equal(reflectx.Indirect(reflect.ValueOf(c.req)).Interface()))
					})
				})

			}
		})
	}
}
