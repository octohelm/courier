package statuserror

import (
	"fmt"
	"net/http"

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

	Status int `json:"-"`
}

func (e *Descriptor) UnmarshalErrorResponse(statusCode int, raw []byte) error {
	d := Descriptor{}
	d.Status = statusCode

	errResp := &ErrorResponse{}
	if err := json.Unmarshal(raw, errResp); err != nil {
		d.Message = string(raw)
		*e = d
		return nil
	}

	switch len(errResp.Errors) {
	case 0:
		d.Message = errResp.Msg
	case 1:
		d = *errResp.Errors[0]
		d.Status = statusCode
	}

	if errResp.Extra != nil {
		if v, ok := errResp.Extra["title"].(string); ok {
			d.Message = v
		}
		if v, ok := errResp.Extra["detail"].(string); ok {
			d.Message = v
		}
	}

	*e = d

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
		er.Code = ERR_CODD_INVALID_PARAMETER
		er.Status = http.StatusBadRequest
		er.Pointer = v.JSONPointer()
	}

	if v, ok := err.(WithErrCode); ok {
		er.Code = v.ErrCode()
	}

	if er.Code == "" {
		er.Code = ErrCodeOf(err)
	}

	if er.Code == "" {
		er.Code = ERR_CODD_INTERNAL_SERVER_ERROR
	}

	if er.Status == 0 {
		er.Status = http.StatusInternalServerError
	}

	return er
}
