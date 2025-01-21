package statuserror

import (
	"net/http"
	"strconv"

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

		ee := asDescriptor(e, source, loc)
		if er == nil {
			er = &ErrorResponse{}

			er.Msg = ee.Message
			er.Code = ee.Status

			if ee.Status == http.StatusBadRequest {
				er.Msg = http.StatusText(ee.Status)
			}
		}

		if len(ee.Errors) > 0 {
			er.Errors = append(er.Errors, ee.Errors...)
		} else {
			er.Errors = append(er.Errors, ee)
		}
	}

	if er == nil {
		return &ErrorResponse{
			Code: http.StatusInternalServerError,
			Msg:  err.Error(),
		}
	}

	return er
}

type ErrorResponse struct {
	// 错误状态码
	Code int `json:"code,omitzero"`
	// 错误信息
	Msg string `json:"msg,omitzero"`
	// 错误详情
	Errors []*Descriptor `json:"errors,omitzero"`
}

func (e *ErrorResponse) StatusCode() int {
	if e.Code > 1000 {
		i, _ := strconv.ParseUint(strconv.FormatUint(uint64(e.Code), 10)[0:3], 10, 64)
		return int(i)
	}
	return e.Code
}

func (e *ErrorResponse) Unwrap() []error {
	if e.Errors != nil {
		return slices.Map(e.Errors, func(e *Descriptor) error {
			return e
		})
	}
	return nil
}
