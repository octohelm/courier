package jsonschema

import (
	"net/url"
	"testing"

	"github.com/octohelm/courier/internal/testingutil"
	"github.com/onsi/gomega"
)

func TestExternalDoc(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		gomega.NewWithT(t).Expect(testingutil.MustJSONRaw(ExternalDoc{})).To(gomega.Equal(`{}`))
	})

	t.Run("with url", func(t *testing.T) {
		gomega.NewWithT(t).Expect(testingutil.MustJSONRaw(ExternalDoc{
			URL: (&url.URL{
				Scheme: "https",
				Host:   "google.com",
			}).String(),
		})).To(gomega.Equal(`{"url":"https://google.com"}`))
	})

	t.Run("with url and description", func(t *testing.T) {
		gomega.NewWithT(t).Expect(testingutil.MustJSONRaw(ExternalDoc{
			URL: (&url.URL{
				Scheme: "https",
				Host:   "google.com",
			}).String(),
			Description: "google",
		})).To(gomega.Equal(`{"description":"google","url":"https://google.com"}`))
	})
}
