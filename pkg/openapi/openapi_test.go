package openapi

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/octohelm/courier/pkg/openapi/jsonschema"
)

func ExampleOpenAPI() {
	openapi := NewOpenAPI()

	openapi.Version = "1.0.0"
	openapi.Title = "Swagger Petstore"
	openapi.License = &License{
		LicenseObject: LicenseObject{
			Name: "MIT",
		},
	}

	openapi.AddTag(nil)
	openapi.AddTag(NewTag("pets"))

	openapi.AddSecurityScheme("token", NewHTTPSecurityScheme("bearer", "JWT"))

	openapi.AddServer(NewServer("http://petstore.swagger.io/v1"))

	openapi.AddSchema("Pet", jsonschema.ObjectOf(jsonschema.Props{
		"id":   jsonschema.Long(),
		"name": jsonschema.String(),
		"tag":  jsonschema.String(),
	}, "id", "name"))

	openapi.AddSchema("Pets", jsonschema.ItemsOf(openapi.RefSchema("Pet")))

	openapi.AddSchema("Error", jsonschema.ObjectOf(jsonschema.Props{
		"code":    jsonschema.Integer(),
		"message": jsonschema.String(),
	}, "code", "message"))

	{
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
			resp.AddContent("application/json", NewMediaTypeWithSchema(openapi.RefSchema("Pets")))

			op.AddResponse(http.StatusOK, resp)
		}

		{
			resp := NewResponse("unexpected error")
			resp.AddContent("application/json", NewMediaTypeWithSchema(openapi.RefSchema("Error")))

			op.SetDefaultResponse(resp)
		}

		openapi.AddOperation(GET, "/pets", op)
	}

	{
		op := NewOperation("createPets")
		op.Summary = "Create a pet"
		op.Tags = []string{"pets"}

		{
			resp := NewResponse("Null response")

			op.AddResponse(http.StatusNoContent, resp)
		}

		{
			resp := NewResponse("unexpected error")
			resp.AddContent("application/json", NewMediaTypeWithSchema(openapi.RefSchema("Error")))

			op.SetDefaultResponse(resp)
		}

		openapi.AddOperation(POST, "/pets", op)
	}

	data, _ := json.MarshalIndent(openapi, "\t", "\t")
	fmt.Println(string(data))
	// Output
}
