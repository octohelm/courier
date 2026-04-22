package openapi

import (
	"context"
	"errors"
	"net/http"
	"reflect"
	"strings"
	"testing"

	. "github.com/octohelm/x/testing/v2"

	"github.com/octohelm/courier/pkg/courier"
	"github.com/octohelm/courier/pkg/courierhttp"
	pkgopenapi "github.com/octohelm/courier/pkg/openapi"
	"github.com/octohelm/courier/pkg/openapi/jsonschema"
	"github.com/octohelm/courier/pkg/statuserror"
)

type scannerRequestBody struct {
	Name string `json:"name"`
}

type scannerResult struct {
	ID uint64 `json:"id"`
}

type scannerBadRequest struct {
	statuserror.BadRequest

	Reason string `json:"reason"`
}

func (e *scannerBadRequest) Error() string {
	return "bad request"
}

type scannerOp struct {
	courierhttp.MethodPost `path:"/api/items/{id}/{scope...}"`

	ID      string              `name:"id" in:"path"`
	Scope   string              `name:"scope" in:"path"`
	Limit   *int                `name:"limit,omitzero" in:"query"`
	Token   string              `name:"X-Token" in:"header"`
	Session string              `name:"session" in:"cookie"`
	Body    *scannerRequestBody `in:"body"`
}

func (*scannerOp) Output(context.Context) (any, error) { return nil, nil }

func (*scannerOp) RuntimeDoc(names ...string) ([]string, bool) {
	if len(names) == 1 {
		return []string{"字段说明: " + names[0]}, true
	}
	return nil, true
}

func (*scannerOp) ResponseStatusCode() int {
	return http.StatusAccepted
}

func (*scannerOp) ResponseContentType() string {
	return "application/custom+json"
}

func (*scannerOp) ResponseContent() any {
	return &scannerResult{}
}

func (*scannerOp) ResponseErrors() []error {
	return []error{
		&scannerBadRequest{},
	}
}

type scannerDescriptorErrOp struct {
	courierhttp.MethodGet `path:"/api/errors"`
}

func (*scannerDescriptorErrOp) Output(context.Context) (any, error) { return nil, nil }

func (*scannerDescriptorErrOp) ResponseContent() any { return nil }

func (*scannerDescriptorErrOp) ResponseErrors() []error {
	return []error{
		&statuserror.Descriptor{Status: http.StatusBadRequest, Code: "BAD"},
	}
}

type scannerNoMethodOp struct{}

func (*scannerNoMethodOp) Output(context.Context) (any, error) { return nil, nil }

type scannerPlainOp struct {
	courierhttp.MethodGet `path:"/api/plain"`
}

func (*scannerPlainOp) Output(context.Context) (any, error) { return nil, nil }

type scannerNilContentOp struct {
	courierhttp.MethodGet `path:"/api/nil"`
}

func (*scannerNilContentOp) Output(context.Context) (any, error) { return nil, nil }
func (*scannerNilContentOp) ResponseContent() any                { return nil }

type scannerContentTypeResp struct{}

func (*scannerContentTypeResp) ContentType() string { return "text/plain" }

type scannerContentTypeOp struct {
	courierhttp.MethodGet `path:"/api/text"`
}

func (*scannerContentTypeOp) Output(context.Context) (any, error) { return nil, nil }
func (*scannerContentTypeOp) ResponseContent() any                { return &scannerContentTypeResp{} }

type scannerMissingInOp struct {
	courierhttp.MethodPost `path:"/api/missing-in"`

	Name string
}

func (*scannerMissingInOp) Output(context.Context) (any, error) { return nil, nil }

func TestScannerHelpersAndFromRouter(t *testing.T) {
	Then(t, "scanner 主链路与辅助方法覆盖关键分支",
		ExpectMust(func() error {
			r := courierhttp.GroupRouter("/").With(
				courier.NewRouter(&scannerOp{}),
				courier.NewRouter(&scannerDescriptorErrOp{}),
			)

			o := FromRouter(r, Naming(func(s string) string {
				return "Named" + reflect.TypeOf(s).Name()
			}))
			if o == nil {
				return errScanner("nil openapi")
			}

			item, ok := o.Paths.Get("/api/items/{id}/{scope}")
			if !ok || item == nil {
				return errScanner("missing patched path")
			}
			op, ok := item.Get("post")
			if !ok || op == nil {
				return errScanner("missing post operation")
			}
			if op.RequestBody == nil {
				return errScanner("missing request body")
			}
			if len(op.Parameters) != 5 {
				return errScanner("unexpected parameter count")
			}
			if op.Responses["202"] == nil || op.Responses["400"] == nil {
				return errScanner("missing responses")
			}

			errResp := op.Responses["400"]
			if errResp == nil {
				return errScanner("missing 400 response")
			}
			if _, ok := errResp.GetExtension("x-status-return-errors"); !ok {
				return errScanner("missing error extension")
			}

			errItem, ok := o.Paths.Get("/api/errors")
			if !ok || errItem == nil {
				return errScanner("missing error path")
			}
			getOp, ok := errItem.Get("get")
			if !ok || getOp == nil || getOp.Responses["400"] == nil {
				return errScanner("missing descriptor error response")
			}
			return nil
		}),
		ExpectMust(func() error {
			r := courierhttp.GroupRouter("/").With(
				courier.NewRouter(&scannerOp{}),
			)
			o1 := DefaultBuildFunc(r)
			o2 := DefaultBuildFunc(r)
			if o1 != o2 {
				return errScanner("expected cached openapi instance")
			}

			o3 := FromRouter(r)
			if o3 == nil || len(o3.ComponentsObject.Schemas) == 0 {
				return errScanner("expected default naming to register schemas")
			}
			return nil
		}),
		ExpectMust(func() error {
			b := &scanner{
				o: pkgopenapi.NewOpenAPI(),
				opt: buildOption{
					naming: func(s string) string { return "Named" + s },
				},
			}

			if b.Record("demo.Type") {
				return errScanner("first record should be false")
			}
			if !b.Record("demo.Type") {
				return errScanner("second record should be true")
			}

			if b.RefString("demo.Type") != "#/components/schemas/Nameddemo.Type" {
				return errScanner("unexpected ref string")
			}

			b.RegisterSchema("#/components/schemas/Demo", jsonschema.String())
			b.RegisterSchema("#/components/schemas/Demo", jsonschema.String())
			if len(b.o.ComponentsObject.Schemas) != 1 {
				return errScanner("unexpected schema registry size")
			}

			op := pkgopenapi.NewOperation("scan")
			b.scanResponseError(context.Background(), op, courier.NewOperatorFactory(&scannerDescriptorErrOp{}, true))
			b.scanResponseError(context.Background(), op, courier.NewOperatorFactory(&scannerDescriptorErrOp{}, true))
			resp := op.Responses["400"]
			if resp == nil {
				return errScanner("missing repeated error response")
			}
			v, ok := resp.GetExtension("x-status-return-errors")
			if !ok || len(v.([]string)) < 2 {
				return errScanner("missing merged return errors")
			}

			b.scanParameterOrRequestBody(context.Background(), op, reflect.TypeFor[scannerOp]())
			if op.RequestBody == nil || len(op.Parameters) == 0 {
				return errScanner("missing scanned params or body")
			}
			return nil
		}),
		ExpectMust(func() error {
			b := &scanner{
				o: pkgopenapi.NewOpenAPI(),
				opt: buildOption{
					naming: func(s string) string { return "Named" + s },
				},
			}
			op := pkgopenapi.NewOperation("empty")
			b.scanResponse(context.Background(), op, courier.NewOperatorFactory(&scannerNoMethodOp{}, true))
			if len(op.Responses) != 0 {
				return errScanner("unexpected response for operator without method")
			}
			return nil
		}),
		ExpectMust(func() error {
			b := &scanner{
				o: pkgopenapi.NewOpenAPI(),
				opt: buildOption{
					naming: func(s string) string { return "Named" + s },
				},
			}

			plain := pkgopenapi.NewOperation("plain")
			b.scanResponse(context.Background(), plain, courier.NewOperatorFactory(&scannerPlainOp{}, true))
			if plain.Responses["200"] == nil {
				return errScanner("missing default get response")
			}

			nilContent := pkgopenapi.NewOperation("nil")
			b.scanResponse(context.Background(), nilContent, courier.NewOperatorFactory(&scannerNilContentOp{}, true))
			if nilContent.Responses["204"] == nil {
				return errScanner("missing no content response")
			}

			text := pkgopenapi.NewOperation("text")
			b.scanResponse(context.Background(), text, courier.NewOperatorFactory(&scannerContentTypeOp{}, true))
			resp := text.Responses["200"]
			if resp == nil || resp.Content["text/plain"] == nil {
				return errScanner("missing content type describer response")
			}
			return nil
		}),
	)
}

func TestScannerErrorMessages(t *testing.T) {
	t.Run("字段缺少 in 标记时应返回中文上下文", func(t *testing.T) {
		b := &scanner{
			o: pkgopenapi.NewOpenAPI(),
			opt: buildOption{
				naming: func(s string) string { return "Named" + s },
			},
		}
		op := pkgopenapi.NewOperation("MissingIn")

		err := captureScannerPanic(func() {
			b.scanParameterOrRequestBody(context.Background(), op, reflect.TypeFor[scannerMissingInOp]())
		})
		if err == nil {
			t.Fatalf("expected panic error")
		}
		if !strings.Contains(err.Error(), "操作 MissingIn 的字段 Name 缺少 in 标记") {
			t.Fatalf("unexpected error message: %v", err)
		}
	})

	t.Run("FromRouter 应补扫描 OpenAPI 路由失败上下文", func(t *testing.T) {
		r := courierhttp.GroupRouter("/").With(
			courier.NewRouter(&scannerMissingInOp{}),
		)

		err := captureScannerPanic(func() {
			_ = FromRouter(r)
		})
		if err == nil {
			t.Fatalf("expected panic error")
		}
		if !strings.Contains(err.Error(), "扫描 OpenAPI 路由失败") {
			t.Fatalf("unexpected wrapped error message: %v", err)
		}
		if !strings.Contains(err.Error(), "缺少 in 标记") {
			t.Fatalf("expected inner error context, got: %v", err)
		}
	})
}

func captureScannerPanic(fn func()) (err error) {
	defer func() {
		if x := recover(); x != nil {
			switch e := x.(type) {
			case error:
				err = e
			default:
				err = errors.New("unexpected non-error panic")
			}
		}
	}()

	fn()
	return nil
}

func errScanner(msg string) error {
	return errors.New(msg)
}
