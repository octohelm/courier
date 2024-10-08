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

type validationError struct {
}

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

	Type  string
	Value any
}

func (e *ErrInvalidType) Error() string {
	if e.Value == nil {
		return fmt.Sprintf("invalid %s", e.Type)
	}
	return fmt.Sprintf("invalid %s: %s", e.Type, e.Value)
}

type ErrNotMatch struct {
	validationError

	Topic   string
	Current any
	Pattern string
}

func (err *ErrNotMatch) Error() string {
	return fmt.Sprintf("%s should match %v, but got %v", err.Topic, err.Pattern, err.Current)
}

type ErrMultipleOf struct {
	validationError

	Topic      string
	Current    any
	MultipleOf any
}

func (e *ErrMultipleOf) Error() string {
	buf := bytes.NewBuffer(nil)
	buf.WriteString(e.Topic)
	buf.WriteString(fmt.Sprintf(" should be multiple of %v", e.MultipleOf))
	buf.WriteString(fmt.Sprintf(", but got %v", e.Current))
	return buf.String()
}

type NotInEnumError struct {
	validationError

	Topic   string
	Current any
	Enums   []any
}

func (e *NotInEnumError) Error() string {
	buf := bytes.NewBuffer(nil)
	buf.WriteString(e.Topic)
	buf.WriteString(" should be one of ")

	for i, v := range e.Enums {
		if i > 0 {
			buf.WriteString(", ")
		}
		buf.WriteString(fmt.Sprintf("%v", v))
	}

	buf.WriteString(fmt.Sprintf(", but got %v", e.Current))

	return buf.String()
}

type OutOfRangeError struct {
	validationError

	Topic            string
	Current          any
	Minimum          any
	Maximum          any
	ExclusiveMaximum bool
	ExclusiveMinimum bool
}

func (e *OutOfRangeError) Error() string {
	buf := &strings.Builder{}

	buf.WriteString(e.Topic)
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

	buf.WriteString(fmt.Sprintf(", but got %v", e.Current))

	return buf.String()
}
