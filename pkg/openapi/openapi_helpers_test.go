package openapi

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/go-json-experiment/json"
	"github.com/octohelm/courier/pkg/openapi/jsonschema"
	. "github.com/octohelm/x/testing/v2"
)

func TestOpenAPIHelpers(t0 *testing.T) {
	Then(t0, "OpenAPI 辅助对象可正确构建",
		ExpectMust(func() error {
			doc := NewOpenAPI()
			if doc.OpenAPI != "3.1.0" {
				return assertf("unexpected version %s", doc.OpenAPI)
			}

			doc.AddSchema("Pet", jsonschema.ObjectOf(map[string]jsonschema.Schema{
				"id": jsonschema.Integer(),
			}, "id"))

			if doc.RefSchema("Pet") == nil {
				return assertf("expected schema ref")
			}

			op := NewOperation("listPets").
				WithTags("pets").
				WithSummary("list").
				WithDesc("list all pets")

			op.SetRequestBody(&RequestBodyObject{})
			op.AddParameter("limit", InQuery, &Parameter{Schema: jsonschema.Integer()})
			op.AddParameter("ignored", InQuery, nil)
			op.AddResponse(http.StatusOK, &ResponseObject{Description: "ok"})
			op.SetDefaultResponse(&ResponseObject{Description: "default"})

			callback := &PathItemObject{}
			callback.Set("post", NewOperation("notify"))
			op.AddCallback("notify", callback)

			doc.AddOperation(http.MethodGet, "/pets", op)

			pathItem, ok := doc.Paths.Get("/pets")
			if !ok || pathItem == nil {
				return assertf("missing path item")
			}

			getOp, ok := pathItem.Get("get")
			if !ok || getOp == nil {
				return assertf("missing operation")
			}

			if len(getOp.Parameters) != 1 {
				return assertf("unexpected parameter count %d", len(getOp.Parameters))
			}
			if getOp.RequestBody == nil {
				return assertf("missing request body")
			}
			if getOp.Responses["200"] == nil || getOp.Responses["default"] == nil {
				return assertf("missing responses")
			}
			if callbackItem, ok := getOp.Callbacks.Get("notify"); !ok || callbackItem == nil {
				return assertf("missing callback")
			}

			return nil
		}),
		ExpectMust(func() error {
			param := NewParameter("trace", InHeader)
			param.SetDefaultStyle()
			if param.Style != ParameterStyleSimple || param.Explode != nil {
				return assertf("unexpected header style %#v", param)
			}

			queryParam := NewParameter("limit", InQuery)
			queryParam.SetDefaultStyle()
			if queryParam.Style != ParameterStyleForm || queryParam.Explode == nil || !*queryParam.Explode {
				return assertf("unexpected query style %#v", queryParam)
			}

			headers := &HeadersObject{}
			headers.AddHeader("X-Trace", &Parameter{Description: "trace"})
			headers.AddHeader("Ignored", nil)

			content := &ContentObject{}
			content.AddContent("application/json", &MediaTypeObject{Schema: jsonschema.String()})
			content.AddContent("ignored", nil)

			mt := &MediaTypeObject{}
			mt.AddEncoding("items", &EncodingObject{})
			mt.AddEncoding("ignored", nil)

			if len(headers.Headers) != 1 || len(content.Content) != 1 || len(mt.Encoding) != 1 {
				return assertf("unexpected helper sizes")
			}

			return nil
		}),
	)
}

func TestPayloadJSON(t0 *testing.T) {
	Then(t0, "Payload 可进行 JSON 编解码",
		ExpectMust(func() error {
			doc := NewOpenAPI()
			doc.Title = "Demo"

			data, err := json.Marshal(Payload{OpenAPI: *doc})
			if err != nil {
				return err
			}

			var payload Payload
			if err := json.Unmarshal(data, &payload); err != nil {
				return err
			}

			if payload.OpenAPI.OpenAPI != "3.1.0" || payload.Title != "Demo" {
				return assertf("unexpected payload %#v", payload)
			}

			return nil
		}),
	)
}

func assertf(format string, args ...any) error {
	return &helperError{msg: sprintf(format, args...)}
}

type helperError struct {
	msg string
}

func (e *helperError) Error() string {
	return e.msg
}

func sprintf(format string, args ...any) string {
	return fmt.Sprintf(format, args...)
}
