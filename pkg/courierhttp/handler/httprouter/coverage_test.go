package httprouter

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"slices"
	"testing"

	"github.com/juju/ansiterm"
	. "github.com/octohelm/x/testing/v2"

	"github.com/octohelm/courier/internal/pathpattern"
	"github.com/octohelm/courier/internal/request"
	"github.com/octohelm/courier/pkg/courier"
	"github.com/octohelm/courier/pkg/courierhttp"
	"github.com/octohelm/courier/pkg/courierhttp/transport"
	pkgopenapi "github.com/octohelm/courier/pkg/openapi"
)

type openapiViewStub struct{}

func (openapiViewStub) Upgrade(http.ResponseWriter, *http.Request) error { return nil }

type stubRouteHandler struct {
	method      string
	path        string
	pathSegs    pathpattern.Segments
	operators   []*courier.OperatorFactory
	statusCode  int
	description string
}

func (h stubRouteHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	rw.WriteHeader(http.StatusNoContent)
}

func (h stubRouteHandler) OperationID() string                   { return "StubRoute" }
func (h stubRouteHandler) Method() string                        { return h.method }
func (h stubRouteHandler) Path() string                          { return h.path }
func (h stubRouteHandler) PathSegments() request.Segments        { return h.pathSegs }
func (h stubRouteHandler) Summary() string                       { return "stub summary" }
func (h stubRouteHandler) Description() string                   { return h.description }
func (h stubRouteHandler) Deprecated() bool                      { return false }
func (h stubRouteHandler) Operators() []*courier.OperatorFactory { return h.operators }

type stubOp struct {
	courierhttp.MethodGet `path:"/v1/users/{id}"`
}

func (*stubOp) Output(context.Context) (any, error) { return nil, nil }

func TestOpenAPIRelatedHelpers(t *testing.T) {
	Then(t, "OpenAPI 与视图辅助逻辑符合预期",
		ExpectMust(func() error {
			ForbidOpenAPI(true)
			defer ForbidOpenAPI(false)

			_, err := (&OpenAPI{}).Output(context.Background())
			var forbidden *ErrOpenAPIForbidden
			if !errors.As(err, &forbidden) {
				return errHttprouter("expected forbidden error")
			}
			return nil
		}),
		ExpectMust(func() error {
			ops := &operations{oas: pkgopenapi.NewOpenAPI()}
			ctx := courierhttp.OperationInfoProviderInjectContext(context.Background(), ops)

			ret, err := (&OpenAPI{}).Output(ctx)
			if err != nil {
				return err
			}
			payload, ok := ret.(*pkgopenapi.Payload)
			if !ok || payload.OpenAPI.OpenAPI != "3.1.0" {
				return errHttprouter("unexpected openapi payload")
			}
			if (&OpenAPI{}).ResponseContentType() != "application/json" {
				return errHttprouter("unexpected openapi content type")
			}
			return nil
		}),
		ExpectMust(func() error {
			SetOpenAPIViewContents(nil)
			_, err := (&OpenAPIView{}).Output(context.Background())
			if err == nil {
				return errHttprouter("expected openapi view error")
			}

			SetOpenAPIViewContents(openapiViewStub{})
			ret, err := (&OpenAPIView{}).Output(context.Background())
			if err != nil {
				return err
			}
			if _, ok := ret.(transport.Upgrader); !ok {
				return errHttprouter("unexpected openapi view upgrader")
			}
			SetOpenAPIViewContents(nil)
			return nil
		}),
	)
}

func TestGroupAndMuxHelpers(t *testing.T) {
	Then(t, "group、mux 与 operations 辅助分支可正常工作",
		ExpectMust(func() error {
			root := &group{}
			v1 := root.child(pathpattern.Parse("/v1"))
			users := v1.child(pathpattern.Parse("/users"))
			if users.depth() != 2 {
				return errHttprouter("unexpected group depth")
			}
			if users.pathSegments().String() != "/v1/users" {
				return errHttprouter("unexpected path segments")
			}
			if root.String() == "" {
				return errHttprouter("unexpected empty group string")
			}
			return nil
		}),
		ExpectMust(func() error {
			g := &group{
				part: pathpattern.Parse("/{namespace...}"),
				children: map[string]*group{
					"/blobs":     {part: pathpattern.Parse("/blobs")},
					"/manifests": {part: pathpattern.Parse("/manifests")},
					"/blobs/x":   {part: pathpattern.Parse("/blobs")},
				},
			}
			segs := slices.Collect(g.childSegment())
			slices.Sort(segs)
			if len(segs) != 2 || segs[0] != "blobs" || segs[1] != "manifests" {
				return errHttprouter("unexpected child segments")
			}
			return nil
		}),
		ExpectMust(func() error {
			ops := &operations{oas: pkgopenapi.NewOpenAPI()}
			if _, ok := ops.GetOperation("missing"); ok {
				return errHttprouter("unexpected operation hit")
			}
			info := &courierhttp.OperationInfo{ID: "GetUser"}
			ops.add(info)
			if got, ok := ops.GetOperation("GetUser"); !ok || got.ID != "GetUser" {
				return errHttprouter("unexpected operation")
			}
			if ops.OpenAPI() == nil {
				return errHttprouter("missing openapi doc")
			}
			return nil
		}),
		ExpectMust(func() error {
			base := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
				if req.Method != http.MethodDelete {
					http.Error(rw, req.Method, http.StatusBadRequest)
					return
				}
				rw.WriteHeader(http.StatusNoContent)
			})
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/users/1", nil)
			req.Header.Set("X-HTTP-Method-Override", http.MethodDelete)
			methodOverride(base).ServeHTTP(rec, req)
			if rec.Code != http.StatusNoContent {
				return errHttprouter("unexpected method override result")
			}
			return nil
		}),
		ExpectMust(func() error {
			m := &mux{
				operations: &operations{oas: pkgopenapi.NewOpenAPI()},
				w:          ansiterm.NewTabWriter(io.Discard, 0, 4, 2, ' ', 0),
			}
			h := stubRouteHandler{
				method:    http.MethodGet,
				path:      "/v1/users/{id}",
				pathSegs:  pathpattern.Parse("/v1/users/{id}"),
				operators: []*courier.OperatorFactory{courier.NewOperatorFactory(&stubOp{}, true)},
			}

			s := http.NewServeMux()
			m.addHandler(s, h)

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/v1/users/1", nil)
			s.ServeHTTP(rec, req)
			if rec.Code != http.StatusNoContent {
				return errHttprouter("unexpected addHandler response")
			}
			if rec.Header().Get("Server") == "" {
				return errHttprouter("missing server header")
			}

			if colorFmtForMethod(http.MethodGet) == colorFmtForMethod("UNKNOWN") {
				return errHttprouter("unexpected color formatter mapping")
			}

			buf := bytes.NewBuffer(nil)
			if _, err := colorFormatter(0).Fprint(buf, "%s", "demo"); err != nil || buf.String() != "demo" {
				return errHttprouter("unexpected color formatter output")
			}
			return nil
		}),
		ExpectMust(func() error {
			if (&ErrOpenAPIForbidden{}).Error() != "openapi is forbidden" {
				return errHttprouter("unexpected forbidden error text")
			}
			if (&ErrOpenAPIForbidden{}).StatusCode() != http.StatusForbidden {
				return errHttprouter("unexpected forbidden status")
			}
			return nil
		}),
	)
}

func errHttprouter(msg string) error {
	return fmt.Errorf("httprouter test: %w", errors.New(msg))
}
