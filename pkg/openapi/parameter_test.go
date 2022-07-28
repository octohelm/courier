package openapi

import (
	"testing"

	"github.com/octohelm/courier/internal/testingutil"
	"github.com/octohelm/courier/pkg/openapi/jsonschema"
	"github.com/onsi/gomega"
)

func TestParameter(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		actual := Parameter{}
		expected := `{"name":"","in":""}`

		gomega.NewWithT(t).Expect(testingutil.MustJSONRaw(actual)).To(gomega.Equal(expected))
	})

	t.Run("query parameter", func(t *testing.T) {
		actual := QueryParameter("key", jsonschema.String(), true)
		expected := `{"name":"key","in":"query","required":true,"schema":{"type":"string"}}`

		gomega.NewWithT(t).Expect(testingutil.MustJSONRaw(actual)).To(gomega.Equal(expected))
	})

	t.Run("header parameter", func(t *testing.T) {
		actual := HeaderParameter("key", jsonschema.String(), true)
		expected := `{"name":"key","in":"header","required":true,"schema":{"type":"string"}}`

		gomega.NewWithT(t).Expect(testingutil.MustJSONRaw(actual)).To(gomega.Equal(expected))
	})

	t.Run("cookie parameter", func(t *testing.T) {
		actual := CookieParameter("key", jsonschema.String(), true)
		expected := `{"name":"key","in":"cookie","required":true,"schema":{"type":"string"}}`

		gomega.NewWithT(t).Expect(testingutil.MustJSONRaw(actual)).To(gomega.Equal(expected))
	})

	t.Run("path parameter", func(t *testing.T) {
		actual := PathParameter("key", jsonschema.String())
		expected := `{"name":"key","in":"path","required":true,"schema":{"type":"string"}}`

		gomega.NewWithT(t).Expect(testingutil.MustJSONRaw(actual)).To(gomega.Equal(expected))
	})
}
