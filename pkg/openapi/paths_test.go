package openapi

import (
	"net/http"
	"testing"

	"github.com/octohelm/courier/internal/testingutil"
	"github.com/octohelm/courier/pkg/openapi/jsonschema"
	"github.com/onsi/gomega"
)

func TestPaths(t *testing.T) {
	paths := &Paths{}

	op := NewOperation("listPets")
	op.Summary = "List all pets"
	op.Tags = []string{"pets"}

	parameterLimit := QueryParameter("limit", jsonschema.Integer(), false).
		WithDesc("How many items to return at one time (max 100)")

	op.AddParameter(parameterLimit)

	{
		resp := NewResponse("An paged array of pets")

		s := jsonschema.String()
		s.Description = "A link to the next page of responses"
		resp.AddHeader("x-next", NewHeaderWithSchema(s))
		resp.AddContent("application/json", NewMediaTypeWithSchema(jsonschema.String()))

		op.AddResponse(http.StatusOK, resp)
	}

	{
		resp := NewResponse("unexpected error")
		resp.AddContent("application/json", NewMediaTypeWithSchema(jsonschema.String()))

		op.SetDefaultResponse(resp)
	}

	paths.AddOperation(GET, "/pets", op)

	expected := `{"/pets":{"get":{"tags":["pets"],"summary":"List all pets","operationId":"listPets","parameters":[{"name":"limit","in":"query","description":"How many items to return at one time (max 100)","schema":{"type":"integer","format":"int32"}}],"responses":{"200":{"description":"An paged array of pets","headers":{"x-next":{"schema":{"type":"string","description":"A link to the next page of responses"}}},"content":{"application/json":{"schema":{"type":"string"}}}},"default":{"description":"unexpected error","content":{"application/json":{"schema":{"type":"string"}}}}}}}}`

	gomega.NewWithT(t).Expect(testingutil.MustJSONRaw(paths)).To(gomega.Equal(expected))
}
