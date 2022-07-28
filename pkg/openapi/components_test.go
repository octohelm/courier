package openapi

import (
	"testing"

	"github.com/octohelm/courier/internal/testingutil"
	"github.com/octohelm/courier/pkg/openapi/jsonschema"
	"github.com/onsi/gomega"
)

func TestComponents(t *testing.T) {
	components := &Components{}

	components.AddParameter("key", QueryParameter("key", jsonschema.String(), false))
	components.AddParameter("nothing", nil)

	components.AddHeader("key", NewHeaderWithSchema(jsonschema.String()))
	components.AddHeader("nothing", nil)

	components.AddSecurityScheme("key", NewAPIKeySecurityScheme("AccessToken", PositionHeader))
	components.AddSecurityScheme("nothing", nil)

	components.AddLink("key", NewLink("link"))
	components.AddLink("nothing", nil)

	components.AddCallback("key", NewCallback(POST, "callback", NewOperation("op")))
	components.AddCallback("nothing", nil)

	components.AddResponse("key", NewResponse("ok"))
	components.AddResponse("nothing", nil)

	components.AddSchema("key", jsonschema.String())
	components.AddSchema("nothing", nil)

	components.AddExample("key", NewExample())
	components.AddExample("nothing", nil)

	components.AddRequestBody("key", NewRequestBody("desc", false))
	components.AddRequestBody("nothing", nil)

	gomega.NewWithT(t).Expect(testingutil.MustJSONRaw(components)).To(gomega.Equal(`{"schemas":{"key":{"type":"string"}},"responses":{"key":{"description":"ok"}},"parameters":{"key":{"name":"key","in":"query","schema":{"type":"string"}}},"examples":{"key":{}},"requestBodies":{"key":{"description":"desc"}},"headers":{"key":{"schema":{"type":"string"}}},"securitySchemes":{"key":{"type":"apiKey","name":"AccessToken","in":"header"}},"links":{"key":{"operationId":"link"}},"callbacks":{"key":{"callback":{"post":{"operationId":"op","responses":{}}}}}}`))
	gomega.NewWithT(t).Expect(testingutil.MustJSONRaw(components.RefParameter("key"))).To(gomega.Equal(`{"$ref":"#/components/parameters/key"}`))
	gomega.NewWithT(t).Expect(testingutil.MustJSONRaw(components.RefHeader("key"))).To(gomega.Equal(`{"$ref":"#/components/headers/key"}`))
	gomega.NewWithT(t).Expect(testingutil.MustJSONRaw(components.RefLink("key"))).To(gomega.Equal(`{"$ref":"#/components/links/key"}`))
	gomega.NewWithT(t).Expect(testingutil.MustJSONRaw(components.RefCallback("key"))).To(gomega.Equal(`{"$ref":"#/components/callbacks/key"}`))
	gomega.NewWithT(t).Expect(testingutil.MustJSONRaw(components.RefResponse("key"))).To(gomega.Equal(`{"$ref":"#/components/responses/key"}`))
	gomega.NewWithT(t).Expect(testingutil.MustJSONRaw(components.RefSchema("key"))).To(gomega.Equal(`{"$ref":"#/components/schemas/key"}`))
	gomega.NewWithT(t).Expect(testingutil.MustJSONRaw(components.RefExample("key"))).To(gomega.Equal(`{"$ref":"#/components/examples/key"}`))
	gomega.NewWithT(t).Expect(testingutil.MustJSONRaw(components.RefRequestBody("key"))).To(gomega.Equal(`{"$ref":"#/components/requestBodies/key"}`))
	gomega.NewWithT(t).Expect(testingutil.MustJSONRaw(components.RequireSecurity("key"))).To(gomega.Equal(`{"key":[]}`))
	gomega.NewWithT(t).Expect(components.RefParameter("not_found")).To(gomega.BeNil())
	gomega.NewWithT(t).Expect(components.RefHeader("not_found")).To(gomega.BeNil())
	gomega.NewWithT(t).Expect(components.RefLink("not_found")).To(gomega.BeNil())
	gomega.NewWithT(t).Expect(components.RefCallback("not_found")).To(gomega.BeNil())
	gomega.NewWithT(t).Expect(components.RefResponse("not_found")).To(gomega.BeNil())
	gomega.NewWithT(t).Expect(components.RefSchema("not_found")).To(gomega.BeNil())
	gomega.NewWithT(t).Expect(components.RefExample("not_found")).To(gomega.BeNil())
	gomega.NewWithT(t).Expect(components.RefRequestBody("not_found")).To(gomega.BeNil())
}
