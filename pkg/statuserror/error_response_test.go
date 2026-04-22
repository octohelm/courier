package statuserror

import (
	"bytes"
	stdjson "encoding/json"
	"errors"
	"net/http"
	"testing"

	"github.com/go-json-experiment/json"
	. "github.com/octohelm/x/testing/v2"
)

func TestAsErrorResponse(t *testing.T) {
	simpleErr := errors.New("test error")
	statusCodeWrapError := Wrap(simpleErr, http.StatusConflict, "Conflict")

	Then(t, "错误响应可稳定导出为快照",
		ExpectMustValue(func() (Snapshot, error) {
			simpleCompact, err := marshalErrorResponseForSnapshot(AsErrorResponse(simpleErr, "x@v1"))
			if err != nil {
				return nil, err
			}
			simpleBuf := bytes.NewBuffer(nil)
			if err := stdjson.Indent(simpleBuf, simpleCompact, "", "  "); err != nil {
				return nil, err
			}

			wrappedCompact, err := marshalErrorResponseForSnapshot(AsErrorResponse(statusCodeWrapError, "x@v1"))
			if err != nil {
				return nil, err
			}
			wrappedBuf := bytes.NewBuffer(nil)
			if err := stdjson.Indent(wrappedBuf, wrappedCompact, "", "  "); err != nil {
				return nil, err
			}

			return SnapshotOf(
				SnapshotFileFromRaw("simple_error.json", simpleBuf.Bytes()),
				SnapshotFileFromRaw("status_code_wrap_error.json", wrappedBuf.Bytes()),
			), nil
		}, MatchSnapshot("all")),
	)
}

func marshalErrorResponseForSnapshot(resp *ErrorResponse) ([]byte, error) {
	return json.Marshal(struct {
		Code   int           `json:"code"`
		Errors []*Descriptor `json:"errors"`
		Msg    string        `json:"msg"`
	}{
		Code:   resp.Code,
		Errors: resp.Errors,
		Msg:    resp.Msg,
	})
}
