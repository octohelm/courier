package openapi

import (
	"testing"

	"github.com/octohelm/courier/internal/testingutil"
	"github.com/octohelm/courier/pkg/openapi/jsonschema"
	"github.com/onsi/gomega"
)

func TestResponse(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		gomega.NewWithT(t).Expect(testingutil.MustJSONRaw(Response{})).To(gomega.Equal(`{"description":""}`))
	})

	t.Run("with header and content and link", func(t *testing.T) {
		resp := NewResponse("desc")
		resp.AddHeader("x-next", NewHeaderWithSchema(jsonschema.String()))
		resp.AddContent("application/json", NewMediaTypeWithSchema(jsonschema.String()))

		link := NewLink("getByUserId")
		link.AddParameter("userId", "$response.body#/id")

		resp.AddLink("GetUserByUserId", link)

		gomega.NewWithT(t).Expect(testingutil.MustJSONRaw(resp)).To(gomega.Equal(`{"description":"desc","headers":{"x-next":{"schema":{"type":"string"}}},"content":{"application/json":{"schema":{"type":"string"}}},"links":{"GetUserByUserId":{"operationId":"getByUserId","parameters":{"userId":"$response.body#/id"}}}}`))
	})
}
