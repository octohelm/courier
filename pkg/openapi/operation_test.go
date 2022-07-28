package openapi

import (
	"net/http"
	"testing"

	"github.com/octohelm/courier/internal/testingutil"
	"github.com/octohelm/courier/pkg/openapi/jsonschema"
	"github.com/onsi/gomega"
)

func TestOperation(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		actual := Operation{}
		expected := `{"operationId":"","responses":{}}`

		gomega.NewWithT(t).Expect(testingutil.MustJSONRaw(actual)).To(gomega.Equal(expected))
	})

	t.Run("full", func(t *testing.T) {
		op := NewOperation("listPets").WithSummary("List all pets").WithDesc("desc").WithTags("pets")

		op.AddParameter(
			QueryParameter("limit", jsonschema.Integer(), true).WithDesc("How many items to return at one time (max 100)"),
		)

		op.AddParameter(
			QueryParameter("size", jsonschema.Integer(), false),
		)

		{
			resp := NewResponse("An paged array of pets")

			s := jsonschema.String()
			s.Description = "A link to the next page of responses"
			resp.AddHeader("x-next", NewHeaderWithSchema(s))
			resp.AddContent("application/json", NewMediaTypeWithSchema(jsonschema.String()))

			op.AddResponse(http.StatusNoContent, nil)
			op.AddResponse(http.StatusOK, resp)
		}

		{
			resp := NewResponse("unexpected error")
			resp.AddContent("text/html", NewMediaTypeWithSchema(jsonschema.String()))

			op.SetDefaultResponse(nil)
			op.SetDefaultResponse(resp)
		}

		opCallback := NewOperation("callback")
		rb := NewRequestBody("", false)
		rb.AddContent("application/json", NewMediaTypeWithSchema(jsonschema.String()))
		opCallback.SetRequestBody(rb)

		opCallback.AddResponse(http.StatusOK, NewResponse("OK"))

		op.AddCallback("myEvent", NewCallback(POST, `{$request.query.callbackUrl}?event={$request.query.event}`, opCallback))

		expected := `{"tags":["pets"],"summary":"List all pets","description":"desc","operationId":"listPets","parameters":[{"name":"limit","in":"query","description":"How many items to return at one time (max 100)","required":true,"schema":{"type":"integer","format":"int32"}},{"name":"size","in":"query","schema":{"type":"integer","format":"int32"}}],"responses":{"200":{"description":"An paged array of pets","headers":{"x-next":{"schema":{"type":"string","description":"A link to the next page of responses"}}},"content":{"application/json":{"schema":{"type":"string"}}}},"default":{"description":"unexpected error","content":{"text/html":{"schema":{"type":"string"}}}}},"callbacks":{"myEvent":{"{$request.query.callbackUrl}?event={$request.query.event}":{"post":{"operationId":"callback","requestBody":{"content":{"application/json":{"schema":{"type":"string"}}}},"responses":{"200":{"description":"OK"}}}}}}}`

		gomega.NewWithT(t).Expect(testingutil.MustJSONRaw(op)).To(gomega.Equal(expected))
	})
}
