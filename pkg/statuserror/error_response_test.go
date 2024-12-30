package statuserror

import (
	"errors"
	"net/http"
	"testing"

	testingx "github.com/octohelm/x/testing"
)

func TestErrorResponse(t *testing.T) {
	simpleErr := errors.New("test error")
	statusCodeWrapError := Wrap(simpleErr, http.StatusConflict, "Conflict")

	s := testingx.NewSnapshot().
		With("simple_error.json", testingx.MustAsJSON(AsErrorResponse(simpleErr, "x@v1"))).
		With("status_code_wrap_error.json", testingx.MustAsJSON(AsErrorResponse(statusCodeWrapError, "x@v1")))

	testingx.Expect(t, s, testingx.MatchSnapshot("all"))
}
