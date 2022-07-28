package jsonschema

import (
	"net/url"
	"testing"

	"github.com/octohelm/courier/internal/testingutil"
	testingx "github.com/octohelm/x/testing"
)

func TestExternalDoc(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		testingx.Expect(t, testingutil.MustJSONRaw(ExternalDoc{}), testingx.Equal(`{}`))
	})

	t.Run("with url", func(t *testing.T) {
		testingx.Expect(t, testingutil.MustJSONRaw(ExternalDoc{
			URL: (&url.URL{
				Scheme: "https",
				Host:   "google.com",
			}).String(),
		}), testingx.Equal(`{"url":"https://google.com"}`))
	})

	t.Run("with url and description", func(t *testing.T) {
		testingx.Expect(t, testingutil.MustJSONRaw(ExternalDoc{
			URL: (&url.URL{
				Scheme: "https",
				Host:   "google.com",
			}).String(),
			Description: "google",
		}), testingx.Equal(`{"description":"google","url":"https://google.com"}`))
	})
}
