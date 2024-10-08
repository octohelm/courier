package statuserror

import (
	"fmt"
	"go/ast"
	"net/http"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/go-json-experiment/json"
	"github.com/octohelm/x/ptr"
	"github.com/octohelm/x/slices"
)

func AsErrorResponse(err error, source string) *ErrorResponse {
	if err == nil {
		return nil
	}

	var er *ErrorResponse

	for e := range All(err) {
		ee := asErrorResponse(e, source)

		if er == nil {
			er = ptr.Ptr(*ee)
		}

		if len(ee.Errors) > 0 {
			er.Errors = append(er.Errors, ee.Errors...)
		} else {
			er.Errors = append(er.Errors, ee)
		}
	}

	return er
}

func asErrorResponse(err error, source string) *ErrorResponse {
	if errorResponse, ok := err.(*ErrorResponse); ok {
		return errorResponse
	}

	er := &ErrorResponse{
		Source: source,
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

	if v, ok := err.(WithLocation); ok {
		in, path := v.Location()

		er.Location = &Location{
			In:   in,
			Path: path,
		}
	}

	if v, ok := err.(WithErrKey); ok {
		er.Key = v.ErrKey()
	}

	if er.Code == 0 {
		er.Code = http.StatusInternalServerError
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

	return er
}

type ErrorResponse struct {
	Code     int              `json:"code"`
	Key      string           `json:"key"`
	Msg      string           `json:"msg"`
	Desc     string           `json:"desc,omitempty"`
	Source   string           `json:"source,omitempty"`
	Location *Location        `json:"location,omitempty"`
	Errors   []*ErrorResponse `json:"errors,omitempty"`

	// Deprecated
	ErrorFields []DeprecatedErrorField `json:"errorFields,omitempty"`
}

type errorResponse ErrorResponse

func (e ErrorResponse) MarshalJSON() ([]byte, error) {
	if len(e.Errors) > 0 {
		errorFields := make([]DeprecatedErrorField, 0, len(e.Errors))
		for _, err := range e.Errors {
			if loc := err.Location; loc != nil {
				errorFields = append(errorFields, DeprecatedErrorField{
					Field: loc.FieldPath(),
					In:    loc.In,
					Msg:   err.Msg,
				})
			}
		}

		e.ErrorFields = errorFields
	}

	return json.Marshal(errorResponse(e))
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

type Location struct {
	In   string `json:"in"`
	Path []any  `json:"path"`
}

func (l Location) FieldPath() string {
	buf := &strings.Builder{}
	for i := 0; i < len(l.Path); i++ {
		switch keyOrIndex := l.Path[i].(type) {
		case string:
			if buf.Len() > 0 {
				buf.WriteRune('.')
			}
			buf.WriteString(keyOrIndex)
		case int:
			buf.WriteString(fmt.Sprintf("[%d]", keyOrIndex))
		}
	}
	return buf.String()
}

func (l Location) String() string {
	return fmt.Sprintf("%s: %s", l.In, l.FieldPath())
}
