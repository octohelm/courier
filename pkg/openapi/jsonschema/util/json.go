package util

import (
	"bytes"
	"strconv"

	"github.com/go-json-experiment/json"
	"github.com/go-json-experiment/json/jsontext"
	"github.com/octohelm/courier/pkg/openapi/jsonschema"
	validatorerrors "github.com/octohelm/courier/pkg/validator"
	"github.com/pkg/errors"
)

func UnmarshalTaggedUnionFromJSON(data []byte, ut jsonschema.GoTaggedUnionType) error {
	dec := jsontext.NewDecoder(bytes.NewReader(data))

	t, err := dec.ReadToken()
	if err != nil {
		return err
	}

	if t.Kind() != '{' {
		return errors.New("tagged union must be an object, starts with `{`")
	}

	discriminatorValue := ""

	for kind := dec.PeekKind(); kind != '}'; kind = dec.PeekKind() {
		k, err := dec.ReadValue()
		if err != nil {
			return err
		}

		var key string
		if err := json.Unmarshal(k, &key); err != nil {
			return err
		}

		v, err := dec.ReadValue()
		if err != nil {
			return err
		}

		if key == ut.Discriminator() {
			if err := json.Unmarshal(v, &discriminatorValue); err != nil {
				return err
			}
			break
		}
	}

	if v, ok := ut.Mapping()[discriminatorValue]; ok {
		dec := jsontext.NewDecoder(bytes.NewReader(data))
		if err := json.UnmarshalDecode(dec, v); err != nil {
			return convertErr(dec, err)
		}
		ut.SetUnderlying(v)
		return nil
	}

	return errors.Errorf("Unsupported Kind %s", discriminatorValue)
}

func keyPathFromJSONPointer(jsonPointer jsontext.Pointer) []any {
	final := make([]any, 0)

	for key := range jsonPointer.Tokens() {
		d, err := strconv.ParseInt(key, 10, 64)
		if err == nil {
			final = append(final, int(d))
		} else {
			final = append(final, key)
		}
	}

	return final
}

func convertErr(dec *jsontext.Decoder, err error) error {
	if err == nil {
		return nil
	}

	var x *json.SemanticError
	if errors.As(err, &x) {
		errSet := validatorerrors.NewErrorSet()
		errSet.AddErr(err, keyPathFromJSONPointer(dec.StackPointer())...)
		return errSet.Err()
	}

	errSet := validatorerrors.NewErrorSet()
	errSet.AddErr(err, keyPathFromJSONPointer(dec.StackPointer())...)
	return errSet.Err()
}
