package openapi

import (
	"testing"

	"github.com/octohelm/courier/internal/testingutil"

	"github.com/octohelm/courier/pkg/openapi/jsonschema"
	"github.com/onsi/gomega"
)

func TestMediaType(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		actual := MediaType{}
		expected := `{}`

		gomega.NewWithT(t).Expect(testingutil.MustJSONRaw(actual)).To(gomega.Equal(expected))
	})

	t.Run("with schema", func(t *testing.T) {
		actual := MediaType{
			MediaTypeObject: MediaTypeObject{
				Schema: jsonschema.String(),
			},
		}
		expected := `{"schema":{"type":"string"}}`

		gomega.NewWithT(t).Expect(testingutil.MustJSONRaw(actual)).To(gomega.Equal(expected))
	})

	t.Run("with schema and example", func(t *testing.T) {
		actual := NewMediaTypeWithSchema(jsonschema.String())
		actual.Example = "some string"

		ex := NewExample()
		ex.Value = "string"
		ex.ExternalValue = "string"

		actual.AddExample("some", ex)

		expected := `{"schema":{"type":"string"},"example":"some string","examples":{"some":{"value":"string","externalValue":"string"}}}`

		gomega.NewWithT(t).Expect(testingutil.MustJSONRaw(actual)).To(gomega.Equal(expected))
	})

	t.Run("with encoding", func(t *testing.T) {
		actual := NewMediaTypeWithSchema(jsonschema.String())

		e := NewEncoding()
		e.ContentType = "application/json"
		e.Style = ParameterStyleSimple

		actual.AddEncoding("utf-8", e)

		expected := `{"schema":{"type":"string"},"encoding":{"utf-8":{"contentType":"application/json","style":"simple"}}}`

		gomega.NewWithT(t).Expect(testingutil.MustJSONRaw(actual)).To(gomega.Equal(expected))
	})
}
