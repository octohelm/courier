package errors

import (
	"errors"
	"fmt"
	"strings"

	"github.com/go-json-experiment/json"
	"github.com/go-json-experiment/json/jsontext"
	"github.com/octohelm/courier/pkg/statuserror"
)

func WrapLocation(err error, location string) error {
	if err == nil || location == "" {
		return nil
	}

	return &errWithLocation{
		location: location,
		err:      err,
	}
}

type errWithLocation struct {
	validationError

	location string
	err      error
}

var _ statuserror.WithLocation = &errWithLocation{}

func (err *errWithLocation) Location() string {
	return err.location
}

func (err *errWithLocation) Unwrap() error {
	return err.err
}

func (err *errWithLocation) Error() string {
	return fmt.Sprintf("%s in %s", err.err, err.location)
}

func Join(errors ...error) error {
	if len(errors) == 0 {
		return nil
	}

	errs := make([]error, 0, len(errors))
	for _, err := range errors {
		if err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) == 0 {
		return nil
	}

	return &errSet{
		errs: errs,
	}
}

type errSet struct {
	validationError

	errs []error
}

func (w *errSet) Error() string {
	b := &strings.Builder{}

	for i, err := range w.errs {
		if i > 0 {
			_, _ = fmt.Fprintf(b, "; %s", err)
		} else {
			_, _ = fmt.Fprintf(b, "%s", err)
		}
	}

	return b.String()
}

func (w *errSet) Unwrap() []error {
	return w.errs
}

func unwrapSemanticError(s *json.SemanticError) *json.SemanticError {
	if s.JSONPointer != "" {
		return s
	}

	if s.Err == nil {
		return s
	}

	serr := &json.SemanticError{}
	if errors.As(s.Err, &serr) {
		return unwrapSemanticError(serr)
	}
	return s
}

func PrefixJSONPointer(err error, pointer jsontext.Pointer) error {
	if err == nil {
		return nil
	}

	if pointer == "" {
		return err
	}

	if es, ok := err.(*errSet); ok {
		errs := make([]error, len(es.errs))
		for i, e := range es.errs {
			errs[i] = PrefixJSONPointer(e, pointer)
		}
		return &errSet{
			errs: errs,
		}
	}

	serr := &json.SemanticError{}
	if errors.As(err, &serr) {
		serr2 := unwrapSemanticError(serr)

		child := serr2.JSONPointer
		serr2.JSONPointer = ""
		err = serr2

		if !(child == "" || child == "/") {
			pointer += child
		}
	} else {
		if v, ok := err.(interface {
			JSONPointer() jsontext.Pointer
			Unwrap() error
		}); ok {
			child := v.JSONPointer()

			if !(child == "" || child == "/") {
				pointer += child
				err = v.Unwrap()
			}
		}
	}

	return &errWithJSONPointer{
		jsonPointer: pointer,
		err:         err,
	}
}

func SuffixJSONPointer(err error, suffix jsontext.Pointer) error {
	if err == nil || suffix == "" {
		return nil
	}

	if v, ok := err.(interface {
		JSONPointer() jsontext.Pointer
		Unwrap() error
	}); ok {
		base := v.JSONPointer()

		if base != "" {
			return &errWithJSONPointer{
				jsonPointer: base + suffix,
				err:         v.Unwrap(),
			}
		}
	}

	return err
}

type errWithJSONPointer struct {
	validationError

	jsonPointer jsontext.Pointer
	err         error
}

func (err *errWithJSONPointer) JSONPointer() jsontext.Pointer {
	return err.jsonPointer
}

func (err *errWithJSONPointer) Unwrap() error {
	return err.err
}

func (err *errWithJSONPointer) Error() string {
	return fmt.Sprintf("%s at %s", err.err, err.jsonPointer)
}
