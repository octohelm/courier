package errors

import (
	"bytes"
	"fmt"
	"github.com/go-json-experiment/json/jsontext"
	"reflect"

	reflectx "github.com/octohelm/x/reflect"
)

func WrapJSONPointer(err error, point jsontext.Pointer) error {
	if err == nil {
		return nil
	}

	return &invalid{
		jsonPointer: point,
		err:         err,
	}
}

type invalid struct {
	jsonPointer jsontext.Pointer
	err         error
}

func (err *invalid) JSONPointer() jsontext.Pointer {
	return err.jsonPointer
}

func (err *invalid) Unwrap() error {
	return err.err
}

func (err *invalid) Error() string {
	return err.err.Error()
}

type MissingRequired struct {
}

func (*MissingRequired) Error() string {
	return "missing required field"
}

type NotMatchError struct {
	Target  string
	Current any
	Pattern string
}

func (err *NotMatchError) Error() string {
	return fmt.Sprintf("%s %s not match %v", err.Target, err.Pattern, err.Current)
}

type MultipleOfError struct {
	Target     string
	Current    any
	MultipleOf any
}

func (e *MultipleOfError) Error() string {
	buf := bytes.NewBuffer(nil)
	buf.WriteString(e.Target)
	buf.WriteString(fmt.Sprintf(" should be multiple of %v", e.MultipleOf))
	buf.WriteString(fmt.Sprintf(", but got invalid value %v", e.Current))
	return buf.String()
}

type NotInEnumError struct {
	Target  string
	Current any
	Enums   []any
}

func (e *NotInEnumError) Error() string {
	buf := bytes.NewBuffer(nil)
	buf.WriteString(e.Target)
	buf.WriteString(" should be one of ")

	for i, v := range e.Enums {
		if i > 0 {
			buf.WriteString(", ")
		}
		buf.WriteString(fmt.Sprintf("%v", v))
	}

	buf.WriteString(fmt.Sprintf(", but got invalid value %v", e.Current))

	return buf.String()
}

type OutOfRangeError struct {
	Target           string
	Current          any
	Minimum          any
	Maximum          any
	ExclusiveMaximum bool
	ExclusiveMinimum bool
}

func (e *OutOfRangeError) Error() string {
	buf := bytes.NewBuffer(nil)
	buf.WriteString(e.Target)
	buf.WriteString(" should be")

	if e.Minimum != nil {
		buf.WriteString(" larger")
		if e.ExclusiveMinimum {
			buf.WriteString(" or equal")
		}

		buf.WriteString(fmt.Sprintf(" than %v", reflectx.Indirect(reflect.ValueOf(e.Minimum)).Interface()))
	}

	if e.Maximum != nil {
		if e.Minimum != nil {
			buf.WriteString(" and")
		}

		buf.WriteString(" less")
		if e.ExclusiveMaximum {
			buf.WriteString(" or equal")
		}

		buf.WriteString(fmt.Sprintf(" than %v", reflectx.Indirect(reflect.ValueOf(e.Maximum)).Interface()))
	}

	buf.WriteString(fmt.Sprintf(", but got invalid value %v", e.Current))

	return buf.String()
}
