package statuserror

import (
	"fmt"
	"go/ast"
	"net/http"
	"path/filepath"
	"reflect"
	"strconv"

	"github.com/go-json-experiment/json"
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
	// 错误编码
	Code string `json:"code,omitzero"`
	// 错误信息
	Message string `json:"message,omitzero"`
	// 错误详情
	Description string `json:"description,omitzero"`
	// 错误参数位置 query, header, path, body 等
	Location string `json:"location,omitzero"`
	// 错误参数 json pointer
	Pointer jsontext.Pointer `json:"pointer,omitzero"`
	// 引起错误的源
	Source string `json:"source,omitzero"`
	// 错误链
	Errors []*Descriptor `json:"errors,omitzero"`

	Extra map[string]any `json:",inline"`

	Status int `json:"-"`
}

func (e *Descriptor) UnmarshalErrorResponse(statusCode int, raw []byte) error {
	if err := json.Unmarshal(raw, e); err != nil {
		e.Status = statusCode
		e.Message = string(raw)
		return nil
	}

	if code, _ := strconv.ParseInt(e.Code, 10, 64); code > 0 {
		e.Status = int(code)
	}

	if e.Message == "" && len(e.Errors) == 1 {
		*e = *e.Errors[0]
	}

	if e.Extra != nil {
		if v, ok := e.Extra["msg"].(string); ok {
			e.Message = v
			delete(e.Extra, "msg")
		}
		if v, ok := e.Extra["title"].(string); ok {
			e.Message = v
			delete(e.Extra, "title")
		}
		if v, ok := e.Extra["detail"].(string); ok {
			e.Description = v
			delete(e.Extra, "detail")
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
