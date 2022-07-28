package statuserror

import (
	"net/http"
	"testing"

	testingx "github.com/octohelm/x/testing"
)

func TestParseStatusErrSummary(t *testing.T) {
	_, err := ParseStatusErrSummary(Wrap(nil, http.StatusInternalServerError, "X").Summary())
	testingx.Expect(t, err, testingx.Be[error](nil))
}
