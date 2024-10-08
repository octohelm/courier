package statuserror

import (
	"errors"
)

type WithStatusCode interface {
	StatusCode() int
}

type WithErrKey interface {
	ErrKey() string
}

type WithLocation interface {
	Location() (string, []any)
}

func Join(errs ...error) error {
	return errors.Join(errs...)
}
