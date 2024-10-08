package statuserror

import (
	"fmt"
)

func Wrap(err error, statusCode int, key string) error {
	if err == nil {
		return nil
	}

	return &statusError{
		statusCode: statusCode,
		key:        key,
		wrapError:  wrapError{err},
	}
}

type statusError struct {
	statusCode int
	key        string

	wrapError
}

var _ WithErrKey = &statusError{}

func (e *statusError) ErrKey() string {
	return e.key
}

var _ WithStatusCode = &statusError{}

func (e *statusError) StatusCode() int {
	return e.statusCode
}

func (e *statusError) Error() string {
	return fmt.Sprintf("%s{ code=%d, msg=%q }", e.key, e.statusCode, e.err)
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
