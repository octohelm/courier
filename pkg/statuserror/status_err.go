package statuserror

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

func IsStatusErr(err error) (*StatusErr, bool) {
	if err == nil {
		return nil, false
	}
	if statusError, ok := err.(StatusError); ok {
		return statusError.StatusErr(), ok
	}
	statusErr, ok := err.(*StatusErr)
	return statusErr, ok
}

func FromErr(err error) *StatusErr {
	if err == nil {
		return nil
	}
	if statusErr, ok := IsStatusErr(err); ok {
		return statusErr
	}
	return Wrap(err, http.StatusInternalServerError, "UnknownError", "未知错误")
}

func Wrap(err error, code int, key string, msgAndDesc ...string) *StatusErr {
	if err == nil {
		err = errors.New(key)
	}

	msg := key

	if len(msgAndDesc) > 0 {
		msg = msgAndDesc[0]
	}

	desc := ""

	if len(msgAndDesc) > 1 {
		desc = strings.Join(msgAndDesc[1:], "\n")
	} else {
		if code == http.StatusInternalServerError {
			desc = fmt.Sprintf("%+v", err)
		} else {
			desc = err.Error()
		}
	}

	s := &StatusErr{
		Key:   key,
		Code:  code,
		Msg:   msg,
		Desc:  desc,
		error: errors.WithStack(err),
	}

	return s
}

type StatusErr struct {
	// http code
	Code int `json:"code" xml:"code"`
	// key of err
	Key string `json:"key" xml:"key"`
	// msg of err
	Msg string `json:"msg" xml:"msg"`
	// desc of err
	Desc string `json:"desc" xml:"desc"`
	// can be task error
	// for client to should error msg to end user
	CanBeTalkError bool `json:"canBeTalkError" xml:"canBeTalkError"`
	// error tracing
	Sources []string `json:"sources" xml:"sources"`
	// error in where fields
	ErrorFields ErrorFields `json:"errorFields" xml:"errorFields"`

	error error
}

func (statusErr *StatusErr) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v':
		if s.Flag('+') {
			e := statusErr.Unwrap()
			if w, ok := e.(WithStackTrace); ok {
				stackTrace := w.StackTrace()
				if len(stackTrace) > 1 {
					_, _ = fmt.Fprintf(s, statusErr.Error()+"%+v", stackTrace[1:])
				}
			}
			return
		}
		fallthrough
	case 's':
		_, _ = io.WriteString(s, statusErr.Error())
	case 'q':
		_, _ = fmt.Fprintf(s, "%q", statusErr.Error())
	}
}

type WithStackTrace interface {
	StackTrace() errors.StackTrace
}

func (statusErr *StatusErr) Unwrap() error {
	return statusErr.error
}

func (statusErr *StatusErr) Summary() string {
	b := bytes.NewBufferString("StatusError{")

	_, _ = fmt.Fprintf(b, "key=%q", statusErr.Key)
	_, _ = fmt.Fprintf(b, ",msg=%q", statusErr.Msg)
	_, _ = fmt.Fprintf(b, ",code=%d", statusErr.Code)

	if statusErr.CanBeTalkError {
		_, _ = fmt.Fprintf(b, ",canBeTalkError")
	}

	b.WriteString("}")

	return b.String()
}

func (statusErr *StatusErr) Is(err error) bool {
	e := FromErr(err)
	if statusErr == nil || e == nil {
		return false
	}
	return e.Key == statusErr.Key && e.Code == statusErr.Code
}

func StatusCodeFromCode(code int) int {
	strCode := fmt.Sprintf("%d", code)
	if len(strCode) < 3 {
		return 0
	}
	statusCode, _ := strconv.Atoi(strCode[:3])
	return statusCode
}

func (statusErr *StatusErr) StatusCode() int {
	return StatusCodeFromCode(statusErr.Code)
}

func (statusErr *StatusErr) Error() string {
	b := bytes.NewBufferString("StatusError{")

	_, _ = fmt.Fprintf(b, "key=%q", statusErr.Key)
	_, _ = fmt.Fprintf(b, ",msg=%q", statusErr.Msg)
	_, _ = fmt.Fprintf(b, ",code=%d", statusErr.Code)

	if statusErr.Desc != "" {
		_, _ = fmt.Fprintf(b, ",desc=%q", statusErr.Desc)
	}

	if len(statusErr.ErrorFields) > 0 {
		_, _ = fmt.Fprintf(b, ",errorFields=%q", statusErr.ErrorFields)
	}

	if sources := statusErr.Sources; len(sources) > 0 {
		_, _ = fmt.Fprintf(b, ",sources=%q", strings.Join(statusErr.Sources, ","))
	}

	if statusErr.CanBeTalkError {
		_, _ = fmt.Fprintf(b, ",canBeTalkError")
	}

	b.WriteString("}")

	return b.String()
}

func (statusErr StatusErr) WithMsg(msg string) *StatusErr {
	statusErr.Msg = msg
	return &statusErr
}

func (statusErr StatusErr) WithDesc(desc string) *StatusErr {
	statusErr.Desc = desc
	return &statusErr
}

func (statusErr StatusErr) AppendSource(sourceName string) *StatusErr {
	length := len(statusErr.Sources)
	if length == 0 || statusErr.Sources[length-1] != sourceName {
		statusErr.Sources = append(statusErr.Sources, sourceName)
	}
	return &statusErr
}

func (statusErr StatusErr) EnableErrTalk() *StatusErr {
	statusErr.CanBeTalkError = true
	return &statusErr
}

func (statusErr StatusErr) DisableErrTalk() *StatusErr {
	statusErr.CanBeTalkError = false
	return &statusErr
}

func (statusErr StatusErr) AppendErrorField(in string, field string, msg string) *StatusErr {
	statusErr.ErrorFields = append(statusErr.ErrorFields, NewErrorField(in, field, msg))
	return &statusErr
}

func (statusErr StatusErr) AppendErrorFields(errorFields ...*ErrorField) *StatusErr {
	statusErr.ErrorFields = append(statusErr.ErrorFields, errorFields...)
	return &statusErr
}
