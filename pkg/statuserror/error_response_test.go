package statuserror

import (
	"bytes"
	"errors"
	"github.com/go-json-experiment/json"
	"github.com/go-json-experiment/json/jsontext"
	"golang.org/x/tools/txtar"
	"net/http"
	"testing"

	testingx "github.com/octohelm/x/testing"
)

func TestErrorResponse(t *testing.T) {
	simpleErr := errors.New("test error")
	statusCodeWrapError := Wrap(simpleErr, http.StatusConflict, "Conflict")
	parameterError := Parameter(simpleErr, "body", "path", "to", "key")

	testingx.Expect(t, &txtar.Archive{
		Files: []txtar.File{
			asJSONArchiveFile("simple_error.json", AsErrorResponse(simpleErr, "x@v1")),
			asJSONArchiveFile("status_code_wrap_error.json", AsErrorResponse(statusCodeWrapError, "x@v1")),
			asJSONArchiveFile("parameter_error.json", AsErrorResponse(parameterError, "x@v1")),
		},
	}, testingx.MatchSnapshot("all"))
}

func asJSONArchiveFile(filename string, v any) txtar.File {
	buf := bytes.NewBuffer(nil)
	enc := jsontext.NewEncoder(buf, jsontext.WithIndent("  "))

	if err := json.MarshalEncode(enc, v); err != nil {
		panic(err)
	}

	return txtar.File{
		Name: filename,
		Data: buf.Bytes(),
	}
}
