package transport

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	. "github.com/octohelm/x/testing/v2"

	"github.com/octohelm/courier/internal/httprequest"
	"github.com/octohelm/courier/pkg/courierhttp"
)

type stubUpgrader struct {
	called bool
	err    error
}

func (u *stubUpgrader) Upgrade(w http.ResponseWriter, r *http.Request) error {
	u.called = true
	return u.err
}

func TestOutgoingTransportHelpers(t0 *testing.T) {
	type req struct {
		courierhttp.MethodPost `path:"/nested/:id"`
		ID                     string `name:"id" in:"path"`
	}

	Then(t0, "outgoing transport 可解析 method/path 并构造请求",
		Expect(resolvePathInTag(reflect.TypeFor[req]()), Equal("/nested/{id}")),
		ExpectMust(func() error {
			ot, err := NewOutgoingTransport(context.Background(), req{ID: "1"})
			if err != nil {
				return err
			}
			if ot.(*outgoingTransport).Method() != http.MethodPost || ot.(*outgoingTransport).Path() != "/nested/{id}" {
				return errTransport("unexpected outgoing transport")
			}
			r, err := ot.NewRequest(context.Background(), req{ID: "1"})
			if err != nil {
				return err
			}
			if r.URL.Path != "/nested/1" {
				return errTransport("unexpected request path: " + r.URL.Path)
			}
			return nil
		}),
	)
}

func TestIncomingTransportWriteResponse(t0 *testing.T) {
	it, _ := NewIncomingTransport(context.Background(), nil)
	req := mustTransportRequest()
	info := httprequest.From(req)
	ctx := courierhttp.OperationInfoInjectContext(context.Background(), &courierhttp.OperationInfo{
		Server: courierhttp.Server{Name: "test"},
	})

	Then(t0, "incoming transport 可写普通响应、错误响应和 upgrade 响应",
		ExpectMust(func() error {
			rec := httptest.NewRecorder()
			it.WriteResponse(ctx, rec, map[string]string{"ok": "1"}, info)
			if rec.Code != http.StatusOK || !bytes.Contains(rec.Body.Bytes(), []byte(`"ok":"1"`)) {
				return errTransport("unexpected normal response")
			}
			return nil
		}),
		ExpectMust(func() error {
			rec := httptest.NewRecorder()
			it.WriteResponse(ctx, rec, errors.New("boom"), info)
			if rec.Code != http.StatusInternalServerError {
				return errTransport("unexpected err response status")
			}
			return nil
		}),
		ExpectMust(func() error {
			rec := httptest.NewRecorder()
			u := &stubUpgrader{}
			it.WriteResponse(ctx, rec, u, info)
			if !u.called {
				return errTransport("upgrader not called")
			}
			return nil
		}),
		ExpectMust(func() error {
			rec := httptest.NewRecorder()
			u := &stubUpgrader{err: errors.New("upgrade failed")}
			it.WriteResponse(ctx, rec, u, info)
			if rec.Code != http.StatusInternalServerError {
				return errTransport("unexpected upgrade err status")
			}
			return nil
		}),
	)
}

func mustTransportRequest() *http.Request {
	req, err := http.NewRequest(http.MethodGet, "http://example.com", io.NopCloser(bytes.NewBuffer(nil)))
	if err != nil {
		panic(err)
	}
	return req
}

func errTransport(msg string) error {
	return &transportErr{msg: msg}
}

type transportErr struct{ msg string }

func (e *transportErr) Error() string { return e.msg }
