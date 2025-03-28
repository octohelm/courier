package errors

import (
	"bytes"
	"fmt"
	"reflect"
	"strings"

	reflectx "github.com/octohelm/x/reflect"
)

func IsValidationError(err error) bool {
	_, ok := err.(ValidationError)
	return ok
}

type ValidationError interface {
	ValidationError()
}

type validationError struct{}

func (validationError) ValidationError() {
}

type ErrMissingRequired struct {
	validationError
}

func (*ErrMissingRequired) Error() string {
	return "missing required field"
}

type ErrInvalidType struct {
	validationError
	Target any

	Type string
}

func (e *ErrInvalidType) Error() string {
	if e.Target == nil {
		return fmt.Sprintf("invalid %s", e.Type)
	}
	return fmt.Sprintf("invalid %s: %s", e.Type, e.Target)
}

type ErrPatternNotMatch struct {
	validationError
	Target any

	Subject string
	Pattern string
	ErrMsg  string
}

func (err *ErrPatternNotMatch) Error() string {
	if err.ErrMsg != "" {
		return fmt.Sprintf("%s %s", err.Subject, err.ErrMsg)
	}
	return fmt.Sprintf("%s should match %v, but got %v", err.Subject, err.Pattern, err.Target)
}

type ErrMultipleOf struct {
	validationError
	Target any

	Subject string

	MultipleOf any
}

func (e *ErrMultipleOf) Error() string {
	buf := bytes.NewBuffer(nil)
	buf.WriteString(e.Subject)
	buf.WriteString(fmt.Sprintf(" should be multiple of %v", e.MultipleOf))
	buf.WriteString(fmt.Sprintf(", but got %v", e.Target))
	return buf.String()
}

type ErrNotInEnum struct {
	validationError
	Target any

	Subject string
	Enums   []any
}

func (e *ErrNotInEnum) Error() string {
	buf := bytes.NewBuffer(nil)
	buf.WriteString(e.Subject)
	buf.WriteString(" should be one of ")

	for i, v := range e.Enums {
		if i > 0 {
			buf.WriteString(", ")
		}
		buf.WriteString(fmt.Sprintf("%v", v))
	}

	buf.WriteString(fmt.Sprintf(", but got %v", e.Target))

	return buf.String()
}

type ErrOutOfRange struct {
	validationError
	Target any

	Subject string

	Minimum          any
	Maximum          any
	ExclusiveMaximum bool
	ExclusiveMinimum bool
}

func (e *ErrOutOfRange) Error() string {
	buf := &strings.Builder{}

	buf.WriteString(e.Subject)
	buf.WriteString(" should be")

	if e.Minimum != nil {
		buf.WriteString(" larger")
		if !e.ExclusiveMinimum {
			buf.WriteString(" or equal")
		}

		buf.WriteString(fmt.Sprintf(" than %v", reflectx.Indirect(reflect.ValueOf(e.Minimum)).Interface()))
	}

	if e.Maximum != nil {
		if e.Minimum != nil {
			buf.WriteString(" and")
		}

		buf.WriteString(" less")
		if !e.ExclusiveMaximum {
			buf.WriteString(" or equal")
		}

		buf.WriteString(fmt.Sprintf(" than %v", reflectx.Indirect(reflect.ValueOf(e.Maximum)).Interface()))
	}

	buf.WriteString(fmt.Sprintf(", but got %v", e.Target))

	return buf.String()
}
