package courierhttp

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"testing"

	. "github.com/octohelm/x/testing/v2"

	"github.com/octohelm/courier/internal/httprequest"
	"github.com/octohelm/courier/pkg/courier"
)

type routeStub struct{}

func (routeStub) Method() string { return http.MethodGet }
func (routeStub) Path() string   { return "/users" }

type operationInfoProviderStub struct{}

func (operationInfoProviderStub) GetOperation(id string) (OperationInfo, bool) {
	return OperationInfo{ID: id}, true
}

type stubRequestInfo struct {
	underlying *http.Request
}

func (s stubRequestInfo) Context() context.Context       { return s.underlying.Context() }
func (s stubRequestInfo) Method() string                 { return s.underlying.Method }
func (s stubRequestInfo) Path() string                   { return s.underlying.URL.Path }
func (s stubRequestInfo) Header() http.Header            { return s.underlying.Header }
func (s stubRequestInfo) Values(string, string) []string { return nil }
func (s stubRequestInfo) Body() io.ReadCloser            { return s.underlying.Body }
func (s stubRequestInfo) Underlying() *http.Request      { return s.underlying }

type resultStub struct {
	data []byte
	err  error
}

func (r resultStub) Into(v any) (courier.Metadata, error) {
	if rw, ok := v.(io.Writer); ok {
		_, _ = rw.Write(r.data)
	}
	return nil, r.err
}

type responseWriterStub struct {
	called bool
}

func (r *responseWriterStub) WriteResponse(context.Context, http.ResponseWriter, RequestInfo) error {
	r.called = true
	return nil
}

func TestContextAndRouteHelpers(t0 *testing.T) {
	ctx := context.Background()
	op := &OperationInfo{
		Server: Server{Name: "courier", Version: "1.0.0"},
		ID:     "ListUsers",
		Method: http.MethodGet,
		Route:  "/users",
	}
	req, _ := http.NewRequest(http.MethodGet, "http://example.com/users", nil)

	Then(t0, "上下文与路由辅助行为符合预期",
		Expect(op.UserAgent(), Equal("courier/1.0.0 (ListUsers)")),
		Expect(Server{Name: "courier"}.UserAgent(), Equal("courier")),
		ExpectMust(func() error {
			ctxWithOp := OperationInfoInjectContext(ctx, op)
			gotOp, ok := OperationInfoFromContext(ctxWithOp)
			if !ok || gotOp != op {
				return fmt.Errorf("unexpected operation info %#v", gotOp)
			}

			ctxWithReq := RequestInjectContext(ctxWithOp, req)
			gotReq, ok := RequestFromContext(ctxWithReq)
			if !ok || gotReq != req {
				return fmt.Errorf("unexpected request %#v", gotReq)
			}

			ctxWithRoute := RouteDescriberInjectContext(ctxWithReq, routeStub{})
			gotRoute, ok := RouteDescriberFromContext(ctxWithRoute)
			if !ok || gotRoute.Path() != "/users" {
				return fmt.Errorf("unexpected route %#v", gotRoute)
			}

			return nil
		}),
		Expect(MethodGet{}.Method(), Equal(http.MethodGet)),
		Expect(MethodHead{}.Method(), Equal(http.MethodHead)),
		Expect(MethodPost{}.Method(), Equal(http.MethodPost)),
		Expect(MethodPut{}.Method(), Equal(http.MethodPut)),
		Expect(MethodPatch{}.Method(), Equal(http.MethodPatch)),
		Expect(MethodDelete{}.Method(), Equal(http.MethodDelete)),
		Expect(MethodConnect{}.Method(), Equal(http.MethodConnect)),
		Expect(MethodOptions{}.Method(), Equal(http.MethodOptions)),
		Expect(MethodTrace{}.Method(), Equal(http.MethodTrace)),
		Expect(fmt.Sprint(Group("/users")), Equal("group(/users)")),
		Expect(fmt.Sprint(BasePath("/api")), Equal("basePath(/api)")),
		Expect(GroupRouter("/users").Routes().String(), Equal("group(/users)")),
		Expect(BasePathRouter("/api").Routes().String(), Equal("basePath(/api)")),
		ExpectMust(func() error {
			group := Group("/users")
			base := BasePath("/api")
			if group.(PathDescriber).Path() != "/users" {
				return fmt.Errorf("unexpected group path")
			}
			if base.(BasePathDescriber).BasePath() != "/api" {
				return fmt.Errorf("unexpected base path")
			}
			if err := op.Init(ctx); err != nil {
				return err
			}
			ctxWithProvider := OperationInfoProviderInjectContext(ctx, operationInfoProviderStub{})
			p, ok := OperationInfoProviderFromContext(ctxWithProvider)
			if !ok {
				return errors.New("missing operation info provider")
			}
			got, ok := p.GetOperation("ListUsers")
			if !ok || got.ID != "ListUsers" {
				return fmt.Errorf("unexpected operation %#v", got)
			}
			ctxWithInjected := op.InjectContext(ctx)
			if gotOp, ok := OperationInfoFromContext(ctxWithInjected); !ok || gotOp != op {
				return fmt.Errorf("unexpected injected op %#v", gotOp)
			}
			return nil
		}),
	)
}

func TestWrapAndWriteResponse(t0 *testing.T) {
	req, _ := http.NewRequest(http.MethodGet, "http://example.com/users", nil)
	reqInfo := httprequest.From(req)

	Then(t0, "普通响应会写出状态码、头和正文",
		ExpectMust(func() error {
			rec := httptest.NewRecorder()
			resp := Wrap(map[string]string{"name": "demo"},
				WithStatusCode(http.StatusAccepted),
				WithContentType("application/custom+json"),
				WithMetadata("X-Trace", "trace-1"),
				WithCookies(&http.Cookie{Name: "sid", Value: "1"}),
			)

			if err := resp.(ResponseWriter).WriteResponse(context.Background(), rec, reqInfo); err != nil {
				return err
			}

			if rec.Code != http.StatusAccepted {
				return fmt.Errorf("unexpected status %d", rec.Code)
			}
			if rec.Header().Get("X-Trace") != "trace-1" {
				return fmt.Errorf("unexpected header %s", rec.Header().Get("X-Trace"))
			}
			if rec.Header().Get("Content-Type") != "application/custom+json" {
				return fmt.Errorf("unexpected content-type %s", rec.Header().Get("Content-Type"))
			}
			if !bytes.Contains(rec.Body.Bytes(), []byte(`"name":"demo"`)) {
				return fmt.Errorf("unexpected body %s", rec.Body.String())
			}
			return nil
		}),
		ExpectMust(func() error {
			rec := httptest.NewRecorder()
			resp := Wrap[any](nil)
			if err := resp.(ResponseWriter).WriteResponse(context.Background(), rec, reqInfo); err != nil {
				return err
			}
			if rec.Code != http.StatusNoContent {
				return fmt.Errorf("unexpected status %d", rec.Code)
			}
			return nil
		}),
		ExpectMust(func() error {
			postReq, _ := http.NewRequest(http.MethodPost, "http://example.com/users", nil)
			rec := httptest.NewRecorder()
			resp := Wrap(map[string]string{"name": "demo"})
			if err := resp.(ResponseWriter).WriteResponse(context.Background(), rec, httprequest.From(postReq)); err != nil {
				return err
			}
			if rec.Code != http.StatusCreated {
				return fmt.Errorf("unexpected status %d", rec.Code)
			}
			return nil
		}),
	)
}

func TestResponseSpecialCases(t0 *testing.T) {
	Then(t0, "特殊响应分支可正确处理",
		ExpectMust(func() error {
			req, _ := http.NewRequest(http.MethodGet, "http://example.com/users", nil)
			rec := httptest.NewRecorder()
			location, _ := url.Parse("http://example.com/next")

			if err := Redirect(http.StatusFound, location).(ResponseWriter).WriteResponse(context.Background(), rec, httprequest.From(req)); err != nil {
				return err
			}
			if rec.Code != http.StatusFound {
				return fmt.Errorf("unexpected status %d", rec.Code)
			}
			if rec.Header().Get("Location") != location.String() {
				return fmt.Errorf("unexpected location %s", rec.Header().Get("Location"))
			}
			return nil
		}),
		ExpectMust(func() error {
			req, _ := http.NewRequest(http.MethodGet, "http://example.com/users", nil)
			rec := httptest.NewRecorder()

			resp := Wrap(resultStub{data: []byte("ok")}, WithContentType("text/plain"))
			if err := resp.(ResponseWriter).WriteResponse(context.Background(), rec, httprequest.From(req)); err != nil {
				return err
			}
			if rec.Body.String() != "ok" {
				return fmt.Errorf("unexpected body %q", rec.Body.String())
			}
			return nil
		}),
		ExpectMust(func() error {
			req, _ := http.NewRequest(http.MethodGet, "http://example.com/users", nil)
			rec := httptest.NewRecorder()
			body := &responseWriterStub{}

			if err := Wrap(body).(ResponseWriter).WriteResponse(context.Background(), rec, httprequest.From(req)); err != nil {
				return err
			}
			if !body.called {
				return errors.New("custom response writer not called")
			}
			return nil
		}),
	)
}

func TestWrapErrorAndErrorType(t0 *testing.T) {
	Then(t0, "错误响应包装与错误类型行为符合预期",
		Expect((&ErrContextCanceled{Reason: "client gone"}).StatusCode(), Equal(499)),
		Expect((&ErrContextCanceled{Reason: "client gone"}).Error(), Equal("context canceled: client gone")),
		ExpectMust(func() error {
			err := WrapError(errors.New("boom"), WithStatusCode(http.StatusBadGateway), WithMetadata("X-Trace", "trace-2"))
			errResp, ok := err.(ErrorResponse)
			if !ok {
				return fmt.Errorf("unexpected error response %T", err)
			}
			if errResp.StatusCode() != http.StatusBadGateway {
				return fmt.Errorf("unexpected status %d", errResp.StatusCode())
			}
			if errResp.ContentType() != "" {
				return fmt.Errorf("unexpected content-type %q", errResp.ContentType())
			}
			if errResp.Meta().Get("X-Trace") != "trace-2" {
				return fmt.Errorf("unexpected meta %v", errResp.Meta())
			}
			if errResp.Unwrap().Error() != "boom" {
				return fmt.Errorf("unexpected unwrap %v", errResp.Unwrap())
			}
			if errResp.Error() != "boom" {
				return fmt.Errorf("unexpected error string %q", errResp.Error())
			}
			return nil
		}),
		ExpectMust(func() error {
			resp := Wrap("ok", WithMetadata("Content-Type", "text/plain"))
			typed, ok := resp.(*response[string])
			if !ok {
				return fmt.Errorf("unexpected response type %T", resp)
			}
			if typed.ContentType() != "text/plain" {
				return fmt.Errorf("unexpected content-type %s", typed.ContentType())
			}
			if !reflect.DeepEqual(typed.Meta(), courier.Metadata{"Content-Type": {"text/plain"}}) {
				return fmt.Errorf("unexpected meta %v", typed.Meta())
			}
			return nil
		}),
		ExpectMust(func() error {
			resp := Wrap("ok", WithCookies(&http.Cookie{Name: "sid", Value: "1"}))
			typed, ok := resp.(*response[string])
			if !ok {
				return fmt.Errorf("unexpected response type %T", resp)
			}
			if len(typed.Cookies()) != 1 {
				return fmt.Errorf("unexpected cookies %v", typed.Cookies())
			}
			typed.SetLocation(&url.URL{Path: "/next"})
			if typed.location == nil || typed.location.Path != "/next" {
				return fmt.Errorf("unexpected location %v", typed.location)
			}
			return nil
		}),
	)
}
