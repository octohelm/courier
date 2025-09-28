package internal

import (
	"bytes"
	"cmp"
	"context"
	"fmt"
	"io"
	"iter"
	"reflect"

	"github.com/go-json-experiment/json/jsontext"
	"github.com/octohelm/courier/internal/jsonflags"
	"github.com/octohelm/courier/pkg/validator"
	validatorerrors "github.com/octohelm/courier/pkg/validator/errors"
)

type ParamValue struct {
	reflect.Value
}

func (p *ParamValue) CanMultiple(sf *jsonflags.StructField) bool {
	return !sf.String && (sf.Type.Kind() == reflect.Slice || sf.Type.Kind() == reflect.Array)
}

func (p *ParamValue) Values(sf *jsonflags.StructField) iter.Seq[reflect.Value] {
	if !sf.String {
		if sf.Type.Kind() == reflect.Slice || sf.Type.Kind() == reflect.Array {
			rv := sf.GetOrNewAt(p.Value)

			return func(yield func(v reflect.Value) bool) {
				for i := 0; i < rv.Len(); i++ {
					if !yield(rv.Index(i)) {
						return
					}
				}
			}
		}
	}

	return func(yield func(v reflect.Value) bool) {
		if !yield(sf.GetOrNewAt(p.Value)) {
			return
		}
	}
}

func (p *ParamValue) AddrValues(sf *jsonflags.StructField, n int) iter.Seq2[int, reflect.Value] {
	if p.CanMultiple(sf) {
		if n == 0 {
			return func(yield func(int, reflect.Value) bool) {
			}
		}

		rv := sf.GetOrNewAt(p.Value)

		if rv.Cap() < n {
			rv.Grow(n)
			rv.SetLen(n)
		}

		return func(yield func(i int, v reflect.Value) bool) {
			for i := 0; i < rv.Len(); i++ {
				if !yield(i, rv.Index(i).Addr()) {
					return
				}
			}
		}
	}

	return func(yield func(i int, v reflect.Value) bool) {
		if !yield(0, sf.GetOrNewAt(p.Value).Addr()) {
			return
		}
	}
}

func (p *ParamValue) UnmarshalValues(ctx context.Context, sf *jsonflags.StructField, values []string) error {
	if len(values) == 0 {
		v, err := NewValidatorWithStructField(sf)
		if err != nil {
			return err
		}
		if defaultValuer, ok := v.(validator.WithDefaultValue); ok {
			if defaultValue := defaultValuer.DefaultValue(); len(defaultValue) > 0 {
				if defaultValue.Kind() == '"' {
					unquote, err := jsontext.AppendUnquote(nil, defaultValue)
					if err != nil {
						return err
					}
					defaultValue = unquote
				}
				values = []string{string(defaultValue)}
			}
		}
	}
	readers := make([]io.ReadCloser, len(values))
	for i := range values {
		readers[i] = io.NopCloser(bytes.NewBufferString(values[i]))
	}
	return p.UnmarshalReaders(ctx, sf, readers)
}

func NewValidatorWithStructField(sf *jsonflags.StructField) (validator.Validator, error) {
	opt := validator.Option{}

	opt.Type = sf.Type
	opt.String = sf.String
	opt.Optional = cmp.Or(sf.Omitzero, sf.Omitempty)

	if v, ok := sf.Tag.Lookup("validate"); ok {
		opt.Rule = v
	}

	if v, ok := sf.Tag.Lookup("default"); ok {
		if err := opt.SetDefaultValue(v); err != nil {
			return nil, err
		}
	}

	return validator.New(opt)
}

func (p *ParamValue) UnmarshalReaders(ctx context.Context, sf *jsonflags.StructField, readers []io.ReadCloser) error {
	for i, ptrRv := range p.AddrValues(sf, len(readers)) {
		t, err := New(ptrRv.Elem().Type(), sf.Tag.Get("mime"), "unmarshal")
		if err != nil {
			return err
		}

		if i < len(readers) {
			if err := t.ReadAs(ctx, readers[i], jsonflags.ValueWithStructField(ptrRv, sf)); err != nil {
				if p.CanMultiple(sf) {
					return validatorerrors.PrefixJSONPointer(err, jsontext.Pointer(fmt.Sprintf("/%s/%d", sf.Name, i)))
				}
				return validatorerrors.PrefixJSONPointer(err, jsontext.Pointer(fmt.Sprintf("/%s", sf.Name)))
			}
		} else {
			if !(sf.Omitempty || sf.Omitzero) {
				if p.CanMultiple(sf) {
					return validatorerrors.PrefixJSONPointer(&validatorerrors.ErrMissingRequired{}, jsontext.Pointer(fmt.Sprintf("/%s/%d", sf.Name, i)))
				}
				return validatorerrors.PrefixJSONPointer(&validatorerrors.ErrMissingRequired{}, jsontext.Pointer(fmt.Sprintf("/%s", sf.Name)))
			}
		}

	}
	return nil
}

func (p *ParamValue) MarshalValues(ctx context.Context, sf *jsonflags.StructField) (values []string, err error) {
	for rv := range p.Values(sf) {
		if rv.IsZero() {
			if sf.Omitempty || sf.Omitzero {
				continue
			}
		}

		t, err := New(rv.Type(), sf.Tag.Get("mime"), "marshal")
		if err != nil {
			return nil, err
		}

		b := bytes.NewBuffer(nil)

		if err := func() error {
			ct, err := t.Prepare(ctx, rv)
			if err != nil {
				return err
			}
			defer ct.Close()

			if _, err := io.Copy(b, ct); err != nil {
				return err
			}

			return nil
		}(); err != nil {
			return nil, err
		}

		values = append(values, b.String())
	}

	return values, err
}
