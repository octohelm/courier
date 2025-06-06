package internal

import (
	"errors"
	"reflect"

	"github.com/go-json-experiment/json"
	"github.com/go-json-experiment/json/jsontext"
	"github.com/octohelm/courier/internal/jsonflags"
)

func UnmarshalDecode(dec *jsontext.Decoder, out any, options ...jsontext.Options) error {
	var vv Validator

	if w, ok := out.(jsonflags.Wrapper); ok {
		out = w.Unwrap()

		v, err := NewWithStructField(w.StructField())
		if err != nil {
			return err
		}
		vv = v
	}

	rv, ok := out.(reflect.Value)
	if !ok {
		rv = reflect.ValueOf(out)
	}

	if vv == nil {
		if w, ok := out.(WithStructTagValidate); ok {
			if rule := w.StructTagValidate(); rule != "" {
				v, err := New(ValidatorOption{
					Type: rv.Type(),
					Rule: rule,
				})
				if err != nil {
					return err
				}
				vv = v
			}
		}
	}

	ra := &Pointer{Value: rv, Validator: vv}

	if ra.Kind() != reflect.Ptr {
		return errors.New("unmarshal target must be ptr value")
	}

	return ra.UnmarshalDecode(dec, json.JoinOptions(options...))
}
