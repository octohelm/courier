package statuserror

import (
	"fmt"
)

func Wrap(err error, statusCode int, code string) error {
	if err == nil {
		return nil
	}

	return &statusError{
		statusCode: statusCode,
		code:       code,
		wrapError:  wrapError{err},
	}
}

type statusError struct {
	statusCode int
	code       string

	wrapError
}

var _ WithErrCode = &statusError{}

func (e *statusError) ErrCode() string {
	return e.code
}

var _ WithStatusCode = &statusError{}

func (e *statusError) StatusCode() int {
	return e.statusCode
}

func (e *statusError) Error() string {
	return fmt.Sprintf("%s{message=%q,statusCode=%d}", e.code, e.err, e.statusCode)
}

type wrapError struct {
	err error
}

func (e wrapError) Error() string {
	return e.err.Error()
}

func (e wrapError) Unwrap() error {
	return e.err
}
