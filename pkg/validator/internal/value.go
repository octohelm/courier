package internal

import (
	"bytes"
	"cmp"
	"encoding"
	"fmt"
	validatorerrors "github.com/octohelm/courier/pkg/validator/errors"
	"reflect"

	"github.com/go-json-experiment/json"
	"github.com/go-json-experiment/json/jsontext"
	"github.com/octohelm/courier/pkg/validator/internal/jsonflags"
)

var (
	bytesType       = reflect.TypeFor[[]byte]()
	emptyStructType = reflect.TypeFor[struct{}]()
)

var (
	jsontextValueType     = reflect.TypeFor[jsontext.Value]()
	textUnmarshalerType   = reflect.TypeFor[encoding.TextUnmarshaler]()
	jsonUnmarshalerV1Type = reflect.TypeFor[json.UnmarshalerV1]()
	jsonUnmarshalerV2Type = reflect.TypeFor[json.UnmarshalerV2]()
)

type Value struct {
	reflect.Value
	Option ValidatorOption
}

func (va *Value) typ() reflect.Type {
	typ := va.Value.Type()
	for typ.Kind() == reflect.Pointer {
		typ = typ.Elem()
	}
	return typ
}

func (va *Value) UnmarshalJSONV2(dec *jsontext.Decoder, options json.Options) (err error) {
	typ := va.typ()

	if jsonflags.Implements(typ, jsonUnmarshalerV2Type) || jsonflags.Implements(typ, jsonUnmarshalerV1Type) || jsonflags.Implements(typ, textUnmarshalerType) {
		return (&Primitive{Value: va.Value, Option: va.Option}).UnmarshalJSONV2(dec, options)
	}

	if typ == jsontextValueType || typ == bytesType {
		return (&Primitive{Value: va.Value, Option: va.Option}).UnmarshalJSONV2(dec, options)
	}

	switch typ.Kind() {
	case reflect.Bool:
		return (&Primitive{Value: va.Value, Option: va.Option}).UnmarshalJSONV2(dec, options)
	case reflect.String:
		return (&Primitive{Value: va.Value, Option: va.Option}).UnmarshalJSONV2(dec, options)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return (&Primitive{Value: va.Value, Option: va.Option}).UnmarshalJSONV2(dec, options)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return (&Primitive{Value: va.Value, Option: va.Option}).UnmarshalJSONV2(dec, options)
	case reflect.Float32, reflect.Float64:
		return (&Primitive{Value: va.Value, Option: va.Option}).UnmarshalJSONV2(dec, options)
	case reflect.Struct:
		return (&Struct{Value: va.Value, Option: va.Option}).UnmarshalJSONV2(dec, options)
	case reflect.Map:
		return (&Record{Value: va.Value, Option: va.Option}).UnmarshalJSONV2(dec, options)
	case reflect.Slice, reflect.Array:
		if reflect.SliceOf(va.Type().Elem()).AssignableTo(bytesType) {
			return (&Primitive{Value: va.Value}).UnmarshalJSONV2(dec, options)
		}
		return (&Array{Value: va.Value, Option: va.Option}).UnmarshalJSONV2(dec, options)
	case reflect.Pointer:
		return (&Pointer{Value: va.Value, Option: va.Option}).UnmarshalJSONV2(dec, options)
	case reflect.Interface:
		return (&Any{Value: va.Value, Option: va.Option}).UnmarshalJSONV2(dec, options)
	default:
		return &json.SemanticError{GoType: va.Value.Type()}
	}
}

type Pointer struct {
	reflect.Value

	Option ValidatorOption
}

func (v *Pointer) UnmarshalJSONV2(dec *jsontext.Decoder, options json.Options) error {
	if dec.PeekKind() == 'n' {
		if _, err := dec.ReadToken(); err != nil {
			return err
		}
		v.SetZero()
		return nil
	}

	if v.IsNil() {
		v.Set(reflect.New(v.Type().Elem()))
	}

	return (&Value{Value: v.Elem(), Option: v.Option}).UnmarshalJSONV2(dec, options)
}

type Any struct {
	reflect.Value

	Option ValidatorOption
}

func (v *Any) UnmarshalJSONV2(dec *jsontext.Decoder, options json.Options) error {
	if dec.PeekKind() == 'n' {
		if _, err := dec.ReadToken(); err != nil {
			return err
		}
		v.SetZero()
		return nil
	}

	var value any

	if err := json.UnmarshalDecode(dec, &value, options); err != nil {
		return err
	}

	v.Set(reflect.ValueOf(value))

	return nil
}

type Primitive struct {
	reflect.Value

	Option ValidatorOption
}

func (t *Primitive) UnmarshalJSONV2(dec *jsontext.Decoder, options json.Options) error {
	raw, err := dec.ReadValue()
	if err != nil {
		return err
	}

	validator, err := New(t.Option)
	if err != nil {
		return err
	}

	if validator != nil {
		v, err := validator.Validate(raw)
		if err != nil {
			return validatorerrors.WrapJSONPointer(err, dec.StackPointer())
		}
		raw = v
	}

	if raw == nil {
		return nil
	}

	return json.Unmarshal(raw, t.Value.Addr().Interface(), options)
}

type Record struct {
	reflect.Value

	Option ValidatorOption
}

func (va *Record) UnmarshalJSONV2(dec *jsontext.Decoder, options json.Options) error {
	tok, err := dec.ReadToken()
	if err != nil {
		return err
	}
	k := tok.Kind()
	switch k {
	case 'n':
		va.SetZero()
		return nil
	case '{':
		// init map
		if va.IsNil() {
			va.Set(reflect.MakeMap(va.Type()))
		}

		propName := &Value{Value: reflect.New(va.Type().Key()).Elem()}
		propValue := &Value{Value: reflect.New(va.Type().Elem()).Elem()}

		var seen reflect.Value

		allowDuplicateNames, _ := json.GetOption(options, jsontext.AllowDuplicateNames)
		if !allowDuplicateNames && va.Len() > 0 {
			seen = reflect.MakeMap(reflect.MapOf(propName.Type(), emptyStructType))
		}

		for dec.PeekKind() != '}' {
			// init key
			propName.SetZero()
			// read key
			if err := propName.UnmarshalJSONV2(dec, options); err != nil {
				return err
			}

			if propName.Kind() == reflect.Interface && !propName.IsNil() && !propName.Elem().Type().Comparable() {
				return &json.SemanticError{GoType: va.Type(), Err: fmt.Errorf("invalid incomparable key type %propValue", propName.Elem().Type())}
			}

			// init value
			if v2 := va.MapIndex(propName.Value); v2.IsValid() {
				if !allowDuplicateNames && (!seen.IsValid() || seen.MapIndex(propName.Value).IsValid()) {
					return &json.SemanticError{
						ByteOffset: dec.InputOffset(),
						Err:        fmt.Errorf("duplicate name %s propValue in object", propName.Value.Interface()),
					}
				}

				propValue.Set(v2)
			} else {
				propValue.SetZero()
			}
			// read value
			if err := propValue.UnmarshalJSONV2(dec, options); err != nil {
				return err
			}

			va.SetMapIndex(propName.Value, propValue.Value)
			if seen.IsValid() {
				seen.SetMapIndex(propName.Value, reflect.Zero(emptyStructType))
			}
			if err != nil {
				return err
			}
		}

		// read '}'
		if _, err := dec.ReadToken(); err != nil {
			return err
		}
		return nil
	}

	return &json.SemanticError{JSONKind: k, GoType: va.Type()}
}

type Struct struct {
	reflect.Value

	Option ValidatorOption
}

func (va *Struct) setDefaultIfNeeds(f *jsonflags.StructFields, seen map[string]struct{}, dec *jsontext.Decoder, options jsontext.Options) error {
	for sf := range f.StructField() {
		if _, ok := seen[sf.Name]; ok {
			continue
		}

		validateOption, err := va.getValidateOption(sf)
		if err != nil {
			return err
		}

		dec := jsontext.NewDecoder(bytes.NewBuffer([]byte("null")))
		v := &Value{Value: sf.GetOrNew(va.Value), Option: validateOption}
		if err := v.UnmarshalJSONV2(dec, va.patchJsonOptions(sf, options)); err != nil {
			return err
		}
	}

	return nil
}

func (va *Struct) UnmarshalJSONV2(dec *jsontext.Decoder, options json.Options) error {
	tok, err := dec.ReadToken()
	if err != nil {
		return err
	}
	k := tok.Kind()
	switch k {
	case 'n':
		f, err := jsonflags.StructVisitor.StructFields(va.Type())
		if err != nil {
			return err
		}
		return va.setDefaultIfNeeds(f, map[string]struct{}{}, dec, options)
	case '{':
		f, err := jsonflags.StructVisitor.StructFields(va.Type())
		if err != nil {
			return err
		}

		seen := map[string]struct{}{}

		inlineFallback, hasInlineFallback := f.InlinedFallback()
		var unknown *bytes.Buffer
		var unknownEnc *jsontext.Encoder

		if hasInlineFallback {
			unknown = new(bytes.Buffer)
			unknownEnc = jsontext.NewEncoder(unknown)

			if err := unknownEnc.WriteToken(jsontext.ObjectStart); err != nil {
				return err
			}
		}

		for dec.PeekKind() != '}' {
			propNameTok, err := dec.ReadToken()
			if err != nil {
				return err
			}
			propName := propNameTok.String()

			seen[propName] = struct{}{}

			sf, ok := f.Lookup(propName)
			if !ok {
				if hasInlineFallback {
					propValue, err := dec.ReadValue()
					if err != nil {
						return err
					}
					if err := unknownEnc.WriteToken(jsontext.String(propName)); err != nil {
						return err
					}
					if err := unknownEnc.WriteValue(propValue); err != nil {
						return err
					}
				} else {
					if err := dec.SkipValue(); err != nil {
						return err
					}
				}

				continue
			}

			validateOption, err := va.getValidateOption(sf)
			if err != nil {
				return err
			}
			if err := (&Value{Value: sf.GetOrNew(va.Value), Option: validateOption}).UnmarshalJSONV2(dec, va.patchJsonOptions(sf, options)); err != nil {
				return err
			}
		}

		// read }
		if _, err := dec.ReadToken(); err != nil {
			return err
		}

		if hasInlineFallback {
			if err := unknownEnc.WriteToken(jsontext.ObjectEnd); err != nil {
				return err
			}

			if err := (&Value{Value: inlineFallback.GetOrNew(va.Value)}).UnmarshalJSONV2(jsontext.NewDecoder(unknown), options); err != nil {
				return err
			}
		}

		return va.setDefaultIfNeeds(f, seen, dec, options)
	}

	return &json.SemanticError{JSONKind: k, GoType: va.Type()}
}

func (va *Struct) getValidateOption(sf *jsonflags.StructField) (ValidatorOption, error) {
	opt := ValidatorOption{}

	opt.String = sf.String
	opt.Optional = cmp.Or(sf.Omitzero, sf.Omitempty)

	if v, ok := sf.Tag.Lookup("validate"); ok {
		opt.Rule = v
	}

	if v, ok := sf.Tag.Lookup("default"); ok {
		if err := opt.SetDefaultValue(v); err != nil {
			return ValidatorOption{}, err
		}
	}

	return opt, nil
}

func (va *Struct) patchJsonOptions(sf *jsonflags.StructField, options json.Options) json.Options {
	nextOptions := options
	if sf.String {
		nextOptions = json.JoinOptions(nextOptions, json.StringifyNumbers(true))
	}
	if sf.Format != "" {
		// FIXME until https://github.com/go-json-experiment/json/issues/52
	}
	return nextOptions
}

type Array struct {
	reflect.Value

	Option ValidatorOption
}

func (va *Array) UnmarshalJSONV2(dec *jsontext.Decoder, options json.Options) error {
	tok, err := dec.ReadToken()
	if err != nil {
		return err
	}
	k := tok.Kind()
	switch k {
	case 'n':
		va.SetZero()
		return nil
	case '[':
		mustZero := true
		sliceCap := va.Cap() // array
		if sliceCap > 0 {
			va.SetLen(sliceCap)
		}

		i := 0
		for dec.PeekKind() != ']' {
			if i == sliceCap {
				va.Grow(1)
				sliceCap = va.Cap()
				va.SetLen(sliceCap)
				mustZero = false
			}
			itemValue := &Value{Value: va.Index(i), Option: va.Option}
			i++
			if mustZero {
				itemValue.SetZero()
			}
			if err := itemValue.UnmarshalJSONV2(dec, options); err != nil {
				itemValue.SetLen(i)
				return err
			}
		}

		if i == 0 {
			va.Set(reflect.MakeSlice(va.Type(), 0, 0))
		} else {
			va.SetLen(i)
		}

		// read ]
		if _, err := dec.ReadToken(); err != nil {
			return err
		}

		return nil
	}
	return &json.SemanticError{JSONKind: k, GoType: va.Type()}
}
