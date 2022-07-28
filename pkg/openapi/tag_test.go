package openapi

import (
	"testing"

	"github.com/octohelm/courier/internal/testingutil"

	"github.com/onsi/gomega"
)

func TestTag(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		gomega.NewWithT(t).Expect(testingutil.MustJSONRaw(Tag{})).To(gomega.Equal(`{"name":""}`))
	})
}
