package internal

import (
	"errors"
	"reflect"

	"github.com/go-json-experiment/json"
	"github.com/go-json-experiment/json/jsontext"
	"github.com/octohelm/courier/internal/jsonflags"
)

func UnmarshalDecode(dec *jsontext.Decoder, out any, options ...jsontext.Options) error {
	var validator Validator

	if w, ok := out.(jsonflags.Wrapper); ok {
		out = w.Unwrap()

		v, err := NewWithStructField(w.StructField())
		if err != nil {
			return err
		}
		validator = v
	}

	rv, ok := out.(reflect.Value)
	if !ok {
		rv = reflect.ValueOf(out)
	}

	ra := &Pointer{Value: rv, Validator: validator}

	if ra.Kind() != reflect.Ptr {
		return errors.New("unmarshal target must be ptr value")
	}

	return ra.UnmarshalJSONV2(dec, json.JoinOptions(options...))
}
