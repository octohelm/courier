package statuserror

import (
	"fmt"
	"github.com/go-json-experiment/json"
	"go/ast"
	"net/http"
	"path/filepath"
	"reflect"
	"strconv"

	"github.com/go-json-experiment/json/jsontext"
)

type WithStatusCode interface {
	StatusCode() int
}

type WithErrCode interface {
	ErrCode() string
}

type WithLocation interface {
	Location() string
}

type WithJSONPointer interface {
	JSONPointer() jsontext.Pointer
}

type IntOrString string

type Descriptor struct {
	Code     string           `json:"code,omitzero"`
	Message  string           `json:"message,omitzero"`
	Location string           `json:"location,omitzero"`
	Pointer  jsontext.Pointer `json:"pointer,omitzero"`
	Source   string           `json:"source,omitzero"`

	Status int           `json:"-"`
	Errors []*Descriptor `json:"-"`

	Extra map[string]any `json:",inline"`
}

func (e *Descriptor) UnmarshalErrorResponse(statusCode int, raw []byte) error {
	if err := json.Unmarshal(raw, e); err != nil {
		e.Status = statusCode
		e.Message = string(raw)
		return nil
	}

	v, _ := strconv.ParseInt(e.Code, 10, 64)
	if v > 0 {
		e.Status = 0
	}

	if e.Extra != nil {
		if msg, ok := e.Extra["msg"].(string); ok {
			e.Message = msg
		}
	}

	return nil
}

func (e *Descriptor) StatusCode() int {
	return e.Status
}

func (e *Descriptor) Error() string {
	return fmt.Sprintf("%s{message=%q}", e.Code, e.Message)
}

func asDescriptor(err error, source string, loc string) *Descriptor {
	if errResp, ok := err.(*Descriptor); ok {
		return errResp
	}

	er := &Descriptor{
		Source:   source,
		Location: loc,
	}

	if w, ok := err.(interface{ Unwrap() error }); ok {
		if e := w.Unwrap(); e != nil {
			er.Message = e.Error()
		}
	}

	if er.Message == "" {
		er.Message = err.Error()
	}

	if v, ok := err.(WithStatusCode); ok {
		er.Status = v.StatusCode()
	}

	if v, ok := err.(WithJSONPointer); ok {
		er.Status = http.StatusBadRequest
		er.Code = "InvalidParameter"
		er.Pointer = v.JSONPointer()
	}

	if v, ok := err.(WithErrCode); ok {
		er.Code = v.ErrCode()
	}

	if er.Code == "" {
		rv := reflect.TypeOf(err)
		for rv.Kind() == reflect.Ptr {
			rv = rv.Elem()
		}

		if ast.IsExported(rv.Name()) {
			if p := rv.PkgPath(); p != "" {
				er.Code = filepath.Base(p) + "." + rv.Name()
			} else {
				er.Code = rv.Name()
			}
		}
	}

	if er.Code == "" {
		er.Code = "InternalServerError"
	}

	if er.Status == 0 {
		er.Status = http.StatusInternalServerError
	}

	return er
}
