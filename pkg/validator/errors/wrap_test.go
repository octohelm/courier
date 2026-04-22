package errors

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/go-json-experiment/json"
	"github.com/go-json-experiment/json/jsontext"
	. "github.com/octohelm/x/testing/v2"
)

type pointerWrappedErr struct {
	pointer jsontext.Pointer
	err     error
}

func (e *pointerWrappedErr) JSONPointer() jsontext.Pointer { return e.pointer }
func (e *pointerWrappedErr) Unwrap() error                 { return e.err }
func (e *pointerWrappedErr) Error() string                 { return e.err.Error() }

func TestValidationErrorHelpers(t0 *testing.T) {
	Then(t0, "validation error 与 wrapper 行为符合预期",
		Expect(IsValidationError(&ErrMissingRequired{}), Equal(true)),
		Expect(IsValidationError(fmt.Errorf("plain")), Equal(false)),
		Expect((&ErrInvalidType{Type: "string"}).Error(), Equal("invalid string")),
		Expect((&ErrInvalidType{Type: "string", Target: "1"}).Error(), Equal("invalid string: 1")),
		Expect((&ErrPatternNotMatch{Subject: "name", ErrMsg: "is required"}).Error(), Equal("name is required")),
		ExpectMust(func() error {
			err := WrapLocation(&ErrMissingRequired{}, "query")
			if err == nil {
				return errWrap("expected wrapped location error")
			}
			if err.(interface{ Location() string }).Location() != "query" {
				return errWrap("unexpected location")
			}
			if errors.Unwrap(err).Error() != "missing required field" {
				return errWrap("unexpected unwrap")
			}
			if err.Error() != "missing required field in query" {
				return errWrap("unexpected error text: " + err.Error())
			}
			return nil
		}),
		Expect(Join(), Equal[error](nil)),
		Expect(Join(nil, nil), Equal[error](nil)),
		ExpectMust(func() error {
			err := Join(&ErrMissingRequired{}, &ErrInvalidType{Type: "string", Target: "1"})
			if err == nil {
				return errWrap("expected joined error")
			}
			if err.Error() != "missing required field; invalid string: 1" {
				return errWrap("unexpected joined error: " + err.Error())
			}
			if len(err.(interface{ Unwrap() []error }).Unwrap()) != 2 {
				return errWrap("unexpected unwrap len")
			}
			return nil
		}),
	)
}

func TestJSONPointerHelpers(t0 *testing.T) {
	Then(t0, "JSON pointer 前后缀包装符合预期",
		Expect(PrefixJSONPointer(nil, "/a"), Equal[error](nil)),
		Expect(SuffixJSONPointer(nil, "/a"), Equal[error](nil)),
		ExpectMust(func() error {
			err := SuffixJSONPointer(fmt.Errorf("plain"), "/a")
			if err == nil || err.Error() != "plain" {
				return errWrap("unexpected plain suffix result")
			}
			return nil
		}),
		ExpectMust(func() error {
			err := PrefixJSONPointer(&ErrMissingRequired{}, "/root")
			if err.Error() != "missing required field at /root" {
				return errWrap("unexpected prefixed error: " + err.Error())
			}
			if err.(interface{ JSONPointer() jsontext.Pointer }).JSONPointer() != "/root" {
				return errWrap("unexpected pointer")
			}
			if errors.Unwrap(err).Error() != "missing required field" {
				return errWrap("unexpected unwrap")
			}
			return nil
		}),
		ExpectMust(func() error {
			err := PrefixJSONPointer(Join(&ErrMissingRequired{}, &ErrInvalidType{Type: "string"}), "/root")
			if err == nil || err.Error() != "missing required field at /root; invalid string at /root" {
				return errWrap("unexpected joined prefixed error")
			}
			return nil
		}),
		ExpectMust(func() error {
			err := PrefixJSONPointer(&json.SemanticError{
				Err: &json.SemanticError{
					Err:         &ErrMissingRequired{},
					JSONPointer: "/name",
				},
			}, "/root")
			if !strings.Contains(err.Error(), "missing required field at /root/name") {
				return errWrap("unexpected semantic prefixed error: " + err.Error())
			}
			return nil
		}),
		ExpectMust(func() error {
			err := PrefixJSONPointer(&pointerWrappedErr{
				pointer: "/child",
				err:     &ErrMissingRequired{},
			}, "/root")
			if err.Error() != "missing required field at /root/child" {
				return errWrap("unexpected wrapped prefixed error: " + err.Error())
			}
			return nil
		}),
		ExpectMust(func() error {
			err := SuffixJSONPointer(&pointerWrappedErr{
				pointer: "/root",
				err:     &ErrMissingRequired{},
			}, "/name")
			if err.Error() != "missing required field at /root/name" {
				return errWrap("unexpected suffixed error: " + err.Error())
			}
			return nil
		}),
		ExpectMust(func() error {
			err := SuffixJSONPointer(&pointerWrappedErr{
				pointer: "",
				err:     &ErrMissingRequired{},
			}, "/name")
			if err == nil || err.Error() != "missing required field" {
				return errWrap("expected original error when base pointer empty")
			}
			return nil
		}),
	)
}

func errWrap(msg string) error {
	return &wrapErr{msg: msg}
}

type wrapErr struct {
	msg string
}

func (e *wrapErr) Error() string {
	return e.msg
}
