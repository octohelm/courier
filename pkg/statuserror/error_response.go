package statuserror

import (
	"fmt"
	"go/ast"
	"net/http"
	"path/filepath"
	"reflect"

	"github.com/go-json-experiment/json/jsontext"
	"github.com/octohelm/x/ptr"
	"github.com/octohelm/x/slices"
)

func AsErrorResponse(err error, source string) *ErrorResponse {
	if err == nil {
		return nil
	}

	var er *ErrorResponse

	loc := ""

	for e := range All(err) {
		if ee, ok := e.(WithLocation); ok {
			loc = ee.Location()
			continue
		}

		ee := asErrorResponse(e, source, loc)

		if er == nil {
			er = ptr.Ptr(*ee)

			if er.Code == http.StatusBadRequest {
				er.Pointer = ""
				er.Location = ""
				er.Msg = http.StatusText(er.Code)
			}
		}

		if len(ee.Errors) > 0 {
			er.Errors = append(er.Errors, ee.Errors...)
		} else {
			er.Errors = append(er.Errors, ee)
			ee.Code = 0
		}
	}

	if er == nil {
		return &ErrorResponse{
			Code: http.StatusInternalServerError,
			Key:  "InternalServerError",
			Msg:  err.Error(),
		}
	}

	return er
}

func asErrorResponse(err error, source string, loc string) *ErrorResponse {
	if errResp, ok := err.(*ErrorResponse); ok {
		return errResp
	}

	er := &ErrorResponse{
		Source:   source,
		Location: loc,
	}

	if w, ok := err.(interface{ Unwrap() error }); ok {
		if e := w.Unwrap(); e != nil {
			er.Msg = e.Error()
		}
	}

	if er.Msg == "" {
		er.Msg = err.Error()
	}

	if v, ok := err.(WithStatusCode); ok {
		er.Code = v.StatusCode()
	}

	if v, ok := err.(WithJSONPointer); ok {
		er.Pointer = v.JSONPointer()
		er.Code = http.StatusBadRequest
		er.Key = "InvalidParameter"
	}

	if v, ok := err.(WithErrKey); ok {
		er.Key = v.ErrKey()
	}

	if er.Key == "" {
		rv := reflect.TypeOf(err)
		for rv.Kind() == reflect.Ptr {
			rv = rv.Elem()
		}

		if ast.IsExported(rv.Name()) {
			if p := rv.PkgPath(); p != "" {
				er.Key = filepath.Base(p) + "." + rv.Name()
			} else {
				er.Key = rv.Name()
			}
		}
	}

	if er.Key == "" {
		er.Key = "InternalServerError"
	}

	if er.Code == 0 {
		er.Code = http.StatusInternalServerError
	}

	return er
}

type ErrorResponse struct {
	Code int    `json:"code,omitempty"`
	Key  string `json:"key"`
	Msg  string `json:"msg"`

	Location string           `json:"location,omitzero"`
	Pointer  jsontext.Pointer `json:"pointer,omitzero"`
	Source   string           `json:"source,omitzero"`

	Errors []*ErrorResponse `json:"errors,omitzero"`
}

func (e *ErrorResponse) UnmarshalErrorResponse(statusCode int, raw []byte) error {
	e.Code = statusCode
	e.Key = "UpstreamError"
	e.Msg = string(raw)
	return nil
}

func (e *ErrorResponse) StatusCode() int {
	return e.Code
}

func (e *ErrorResponse) Error() string {
	return fmt.Sprintf("%s{code=%d,msg=%q}", e.Key, e.Code, e.Msg)
}

func (e *ErrorResponse) Unwrap() []error {
	if e.Errors != nil {
		return slices.Map(e.Errors, func(e *ErrorResponse) error {
			return e
		})
	}
	return nil
}
