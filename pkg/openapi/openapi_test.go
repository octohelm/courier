package openapi

import (
	"net/http"
	"testing"

	"github.com/octohelm/courier/internal/testingutil"

	"github.com/octohelm/courier/pkg/openapi/jsonschema"
)

func TestOpenAPI(t *testing.T) {
	openapi := NewOpenAPI()

	openapi.Version = "1.0.0"
	openapi.Title = "Swagger Petstore"
	openapi.License = &License{Name: "MIT"}

	openapi.AddSchema("Pet", jsonschema.ObjectOf(jsonschema.Props{
		"id":   jsonschema.Long(),
		"name": jsonschema.String(),
		"tag":  jsonschema.String(),
	}, "id", "name"))

	openapi.AddSchema("Pets", jsonschema.ArrayOf(openapi.RefSchema("Pet")))

	openapi.AddSchema("Error", jsonschema.ObjectOf(jsonschema.Props{
		"code":    jsonschema.Integer(),
		"message": jsonschema.String(),
	}, "code", "message"))

	{
		op := NewOperation("listPets")

		op.Summary = "List all pets"
		op.Tags = []string{"pets"}

		paramLimit := &Parameter{}
		paramLimit.Schema = jsonschema.Integer()
		paramLimit.Description = "How many items to return at one time (max 100)"
		op.AddParameter("limit", InQuery, paramLimit)

		{
			resp := &ResponseObject{}
			resp.Description = "An paged array of pets"

			s := jsonschema.String()
			s.Description = "A link to the next page of responses"

			resp.AddHeader("x-next", &Parameter{
				Schema: s,
			})

			resp.AddContent("application/json", &MediaTypeObject{
				Schema: openapi.RefSchema("Pets"),
			})

			op.AddResponse(http.StatusOK, resp)
		}

		{
			resp := &ResponseObject{}

			resp.AddContent("application/json", &MediaTypeObject{
				Schema: openapi.RefSchema("Error"),
			})

			op.SetDefaultResponse(resp)
		}

		openapi.AddOperation(http.MethodGet, "/pets", op)
	}

	{
		op := NewOperation("createPets")
		op.Summary = "Create a pet"
		op.Tags = []string{"pets"}

		{
			op.AddResponse(http.StatusNoContent, &ResponseObject{
				Description: "Null response",
			})
		}

		{

			resp := &ResponseObject{}
			resp.Description = "unexpected error"

			resp.AddContent("application/json", &MediaTypeObject{
				Schema: openapi.RefSchema("Error"),
			})

			op.SetDefaultResponse(resp)
		}

		openapi.AddOperation(http.MethodPost, "/pets", op)
	}

	testingutil.PrintJSON(openapi)
}
