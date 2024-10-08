package internal

import (
	"errors"
	"github.com/go-json-experiment/json"
	"github.com/go-json-experiment/json/jsontext"
	"reflect"
)

func UnmarshalDecode(dec *jsontext.Decoder, out any, options ...jsontext.Options) error {
	ra := &Pointer{Value: reflect.ValueOf(out)}

	if ra.Kind() != reflect.Ptr {
		return errors.New("unmarshal target must be ptr value")
	}

	err := json.UnmarshalDecode(dec, ra)
	if err != nil {
		serr := &json.SemanticError{}
		if errors.As(err, &serr) {
			serr.JSONPointer = dec.StackPointer()
			serr.GoType = ra.Type()
		}
		return err
	}

	return nil
}
