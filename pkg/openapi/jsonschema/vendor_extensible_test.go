package jsonschema

import (
	"testing"

	"github.com/octohelm/courier/internal/testingutil"
	"github.com/onsi/gomega"
)

func TestVendorExtensible(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		gomega.NewWithT(t).Expect(testingutil.MustJSONRaw(VendorExtensible{})).To(gomega.Equal(`{}`))
	})

	t.Run("with extensions", func(t *testing.T) {
		e := &VendorExtensible{}
		e.AddExtension("x-b", nil)
		e.AddExtension("x-a", "xxx")

		gomega.NewWithT(t).Expect(testingutil.MustJSONRaw(e)).To(gomega.Equal(`{"x-a":"xxx"}`))
	})
}
