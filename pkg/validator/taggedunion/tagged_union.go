package taggedunion

import (
	"bytes"
	"fmt"

	"github.com/go-json-experiment/json/jsontext"
	"github.com/octohelm/courier/pkg/validator"
	validatorerrors "github.com/octohelm/courier/pkg/validator/errors"
)

type Type interface {
	Discriminator() string
	Mapping() map[string]any
	SetUnderlying(u any)
}

func Unmarshal(data []byte, ut Type) error {
	return UnmarshalDecode(jsontext.NewDecoder(bytes.NewBuffer(data)), ut)
}

func UnmarshalDecode(dec *jsontext.Decoder, ut Type) error {
	t, err := dec.ReadToken()
	if err != nil {
		return err
	}

	switch t.Kind() {
	case 'n':
		return nil
	case '{':
		discriminatorValue := ""

		buf := bytes.NewBuffer(nil)
		enc := jsontext.NewEncoder(buf)
		if err := enc.WriteToken(jsontext.BeginObject); err != nil {
			return err
		}

		for dec.PeekKind() != '}' {
			k, err := dec.ReadToken()
			if err != nil {
				return err
			}
			propName := k.String()

			propValue, err := dec.ReadValue()
			if err != nil {
				return err
			}

			if propName == ut.Discriminator() {
				if err := validator.Unmarshal(propValue, &discriminatorValue); err != nil {
					return err
				}
			}

			if err := enc.WriteToken(jsontext.String(propName)); err != nil {
				return err
			}
			if err := enc.WriteValue(propValue); err != nil {
				return err
			}
		}

		// read }
		if _, err := dec.ReadToken(); err != nil {
			return err
		}

		if err := enc.WriteToken(jsontext.EndObject); err != nil {
			return err
		}

		if v, ok := ut.Mapping()[discriminatorValue]; ok {
			if err := validator.UnmarshalDecode(jsontext.NewDecoder(buf), v); err != nil {
				return validatorerrors.PrefixJSONPointer(err, dec.StackPointer())
			}
			ut.SetUnderlying(v)
			return nil
		}

		if discriminatorValue == "" {
			// when empty discriminatorValue should drop other fields
			return nil
		}

		return validatorerrors.PrefixJSONPointer(
			fmt.Errorf("unsupported %s=%s", ut.Discriminator(), discriminatorValue),
			jsontext.Pointer(fmt.Sprintf("/%s", ut.Discriminator())),
		)
	}

	return &validatorerrors.ErrInvalidType{
		Type: "object",
	}
}
