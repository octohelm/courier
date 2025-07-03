package internal

import (
	"bytes"
	"encoding"
	"errors"
	"fmt"
	"reflect"

	"github.com/go-json-experiment/json"
	"github.com/go-json-experiment/json/jsontext"
	"github.com/octohelm/courier/internal/jsonflags"
	validatorerrors "github.com/octohelm/courier/pkg/validator/errors"
)

var (
	bytesType       = reflect.TypeFor[[]byte]()
	emptyStructType = reflect.TypeFor[struct{}]()
)

var (
	jsontextValueType       = reflect.TypeFor[jsontext.Value]()
	textUnmarshalerType     = reflect.TypeFor[encoding.TextUnmarshaler]()
	jsonUnmarshalerType     = reflect.TypeFor[json.Unmarshaler]()
	jsonUnmarshalerFromType = reflect.TypeFor[json.UnmarshalerFrom]()
)

type Value struct {
	reflect.Value

	Validator Validator
}

func (va *Value) UnmarshalDecode(dec *jsontext.Decoder, options json.Options) (err error) {
	if va.Kind() == reflect.Pointer {
		return (&Pointer{Value: va.Value, Validator: va.Validator}).UnmarshalDecode(dec, options)
	}

	typ := va.Type()

	if reflect.PointerTo(typ).Implements(jsonUnmarshalerFromType) {
		prefix := dec.StackPointer()

		if err := json.UnmarshalDecode(dec, va.Addr().Interface(), options); err != nil {
			return validatorerrors.PrefixJSONPointer(err, prefix)
		}

		return nil
	}

	if reflect.PointerTo(typ).Implements(jsonUnmarshalerType) {
		value, err := dec.ReadValue()
		if err != nil {
			return err
		}

		prefix := dec.StackPointer()

		if err := va.Addr().Interface().(json.Unmarshaler).UnmarshalJSON(value); err != nil {
			return validatorerrors.PrefixJSONPointer(err, prefix)
		}

		return nil
	}

	if reflect.PointerTo(typ).Implements(textUnmarshalerType) {
		return (&Primitive{Value: va.Value, String: true, Validator: va.Validator}).UnmarshalDecode(dec, options)
	}

	if typ == jsontextValueType || typ == bytesType {
		return (&Primitive{Value: va.Value, String: true, Validator: va.Validator}).UnmarshalDecode(dec, options)
	}

	switch typ.Kind() {
	case reflect.String:
		return (&Primitive{Value: va.Value, String: true, Validator: va.Validator}).UnmarshalDecode(dec, options)
	case reflect.Bool:
		return (&Primitive{Value: va.Value, Validator: va.Validator}).UnmarshalDecode(dec, options)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return (&Primitive{Value: va.Value, Validator: va.Validator}).UnmarshalDecode(dec, options)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return (&Primitive{Value: va.Value, Validator: va.Validator}).UnmarshalDecode(dec, options)
	case reflect.Float32, reflect.Float64:
		return (&Primitive{Value: va.Value, Validator: va.Validator}).UnmarshalDecode(dec, options)
	case reflect.Struct:
		return (&Struct{Value: va.Value, Validator: va.Validator}).UnmarshalDecode(dec, options)
	case reflect.Map:
		return (&Record{Value: va.Value, Validator: va.Validator}).UnmarshalDecode(dec, options)
	case reflect.Slice, reflect.Array:
		if reflect.SliceOf(va.Type().Elem()).AssignableTo(bytesType) {
			return (&Primitive{Value: va.Value, String: true, Validator: va.Validator}).UnmarshalDecode(dec, options)
		}
		return (&Array{Value: va.Value, Validator: va.Validator}).UnmarshalDecode(dec, options)
	case reflect.Interface:
		return (&Any{Value: va.Value, Validator: va.Validator}).UnmarshalDecode(dec, options)
	default:
		return &json.SemanticError{GoType: va.Value.Type()}
	}
}

type Pointer struct {
	reflect.Value

	Validator Validator
}

func (va *Pointer) UnmarshalDecode(dec *jsontext.Decoder, options json.Options) error {
	if dec.PeekKind() == 'n' {
		raw, err := dec.ReadValue()
		if err != nil {
			return err
		}

		if validator := va.Validator; validator != nil {
			if err := validator.Validate(raw); err != nil {
				return validatorerrors.PrefixJSONPointer(err, dec.StackPointer())
			}
		}

		if va.CanAddr() {
			va.SetZero()
		}

		return nil
	}

	if va.CanAddr() {
		rv := reflect.New(va.Type().Elem())
		if err := (&Value{Value: rv.Elem(), Validator: va.Validator}).UnmarshalDecode(dec, options); err != nil {
			return err
		}
		va.Set(rv)
		return nil
	}

	if va.IsNil() {
		va.Set(reflect.New(va.Type().Elem()))
	}
	if err := (&Value{Value: va.Elem(), Validator: va.Validator}).UnmarshalDecode(dec, options); err != nil {
		return err
	}
	return nil
}

type Any struct {
	reflect.Value

	Validator Validator
}

func (v *Any) UnmarshalDecode(dec *jsontext.Decoder, options json.Options) error {
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
	String    bool
	Validator Validator
}

func (t *Primitive) UnmarshalDecode(dec *jsontext.Decoder, options json.Options) error {
	tok := dec.PeekKind()
	switch tok {
	case '[':
		stackPointer := dec.StackPointer()

		_, err := dec.ReadToken()
		if err != nil {
			return err
		}

		value, err := dec.ReadValue()
		if err != nil {
			return err
		}

		v := string(value)

		for dec.PeekKind() != ']' {
			if _, err := dec.ReadValue(); err != nil {
				return err
			}
		}

		// read ]
		if _, err := dec.ReadToken(); err != nil {
			return err
		}

		return t.unmarshal([]byte(v), stackPointer, options)
	default:
		value, err := dec.ReadValue()
		if err != nil {
			return err
		}
		return t.unmarshal(value, dec.StackPointer(), options)
	}
}

func (t *Primitive) forceUnquote(value jsontext.Value) (jsontext.Value, error) {
	if value.Kind() == '"' {
		if !t.String {
			return jsontext.AppendUnquote(nil, value)
		}
	}
	return value, nil
}

func (t *Primitive) unmarshal(v jsontext.Value, stackPointer jsontext.Pointer, options json.Options) error {
	value, err := t.forceUnquote(v)
	if err != nil {
		return err
	}

	if validator := t.Validator; validator != nil {
		if err := validator.Validate(value); err != nil {
			return validatorerrors.PrefixJSONPointer(err, stackPointer)
		}
	}

	if err := json.Unmarshal(value, t.Value.Addr().Interface(), options); err != nil {
		serr := &json.SemanticError{}

		if errors.As(err, &serr) {
			serr.JSONPointer = stackPointer
			return serr
		}

		serr.JSONPointer = stackPointer
		serr.Err = err
		return serr
	}
	return nil
}

type Record struct {
	reflect.Value

	Validator Validator
}

func (va *Record) UnmarshalDecode(dec *jsontext.Decoder, options json.Options) error {
	tok, err := dec.ReadToken()
	if err != nil {
		return err
	}
	k := tok.Kind()
	switch k {
	case 'n':
		if validator := va.Validator; validator != nil {
			if err := validator.Validate(jsontext.Value("null")); err != nil {
				return validatorerrors.PrefixJSONPointer(err, dec.StackPointer())
			}
		}

		va.SetZero()
		return nil
	case '{':
		// init map
		if va.IsNil() {
			va.Set(reflect.MakeMap(va.Type()))
		}

		var keyValidator, elemValidator Validator
		if e, ok := va.Validator.(WithKey); ok {
			kRule := e.Key()
			kRule.Type = va.Type().Key()

			v, err := New(kRule)
			if err != nil {
				return &json.SemanticError{
					Err:    err,
					GoType: va.Type(),
				}
			}
			keyValidator = v
		}

		if e, ok := va.Validator.(WithElem); ok {
			elemRule := e.Elem()
			elemRule.Type = va.Type().Elem()

			v, err := New(elemRule)
			if err != nil {
				return &json.SemanticError{
					Err:    err,
					GoType: va.Type(),
				}
			}
			elemValidator = v
		}

		propName := &Value{Value: reflect.New(va.Type().Key()).Elem(), Validator: keyValidator}
		propValue := &Value{Value: reflect.New(va.Type().Elem()).Elem(), Validator: elemValidator}

		var seen reflect.Value

		allowDuplicateNames, _ := json.GetOption(options, jsontext.AllowDuplicateNames)
		if !allowDuplicateNames && va.Len() > 0 {
			seen = reflect.MakeMap(reflect.MapOf(propName.Type(), emptyStructType))
		}

		var errs []error

		for dec.PeekKind() != '}' {
			validKeyValue := true

			// init key
			propName.SetZero()
			// read key
			if keyErr := propName.UnmarshalDecode(dec, options); keyErr != nil {
				if !validatorerrors.IsValidationError(keyErr) {
					return keyErr
				}
				errs = append(errs, validatorerrors.SuffixJSONPointer(keyErr, "/"))

				validKeyValue = false
			}

			if propName.Kind() == reflect.Interface && !propName.IsNil() && !propName.Elem().Type().Comparable() {
				return &json.SemanticError{
					GoType: va.Type(),
					Err:    fmt.Errorf("invalid incomparable key type %propValue", propName.Elem().Type()),
				}
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
			if propErr := propValue.UnmarshalDecode(dec, options); propErr != nil {
				if !validatorerrors.IsValidationError(propErr) {
					return propErr
				}
				errs = append(errs, propErr)
				validKeyValue = false
			}

			if validKeyValue {
				va.SetMapIndex(propName.Value, propValue.Value)

				if seen.IsValid() {
					seen.SetMapIndex(propName.Value, reflect.Zero(emptyStructType))
				}
			}
		}

		// read '}'
		if _, err := dec.ReadToken(); err != nil {
			return err
		}

		if post, ok := va.Validator.(PostValidator); ok {
			if err := post.PostValidate(va.Value); err != nil {
				errs = append(errs, validatorerrors.PrefixJSONPointer(err, dec.StackPointer()))
			}
		}

		return validatorerrors.Join(errs...)
	}

	return &json.SemanticError{JSONKind: k, GoType: va.Type()}
}

type Struct struct {
	reflect.Value

	Validator Validator
}

func (va *Struct) UnmarshalDecode(dec *jsontext.Decoder, options json.Options) error {
	tok, err := dec.ReadToken()
	if err != nil {
		return err
	}
	k := tok.Kind()
	switch k {
	case 'n':
		if validator := va.Validator; validator != nil {
			if err := validator.Validate(jsontext.Value("null")); err != nil {
				return validatorerrors.PrefixJSONPointer(err, dec.StackPointer())
			}
		}
		f, err := jsonflags.Structs.StructFields(va.Type())
		if err != nil {
			return err
		}
		return va.validateRequiredOrSetDefaultIfNeeds(f, map[string]struct{}{}, dec, options)
	case '{':
		f, err := jsonflags.Structs.StructFields(va.Type())
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

			if err := unknownEnc.WriteToken(jsontext.BeginObject); err != nil {
				return err
			}
		}

		var errs []error

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

			validator, err := va.getFieldValidator(sf)
			if err != nil {
				return err
			}

			if err := (&Value{Value: sf.GetOrNewAt(va.Value), Validator: validator}).UnmarshalDecode(dec, va.patchJsonOptions(sf, options)); err != nil {
				if !validatorerrors.IsValidationError(err) {
					return err
				}
				errs = append(errs, err)
			}
		}

		// read }
		if _, err := dec.ReadToken(); err != nil {
			return err
		}

		if hasInlineFallback {
			if err := unknownEnc.WriteToken(jsontext.EndObject); err != nil {
				return err
			}

			if err := (&Value{Value: inlineFallback.GetOrNewAt(va.Value)}).UnmarshalDecode(jsontext.NewDecoder(unknown), options); err != nil {
				if validatorerrors.IsValidationError(err) {
					errs = append(errs, err)
				}
				return err
			}
		}

		if err := va.validateRequiredOrSetDefaultIfNeeds(f, seen, dec, options); err != nil {
			errs = append(errs, err)
		}

		return validatorerrors.Join(errs...)
	}

	return &json.SemanticError{JSONKind: k, GoType: va.Type()}
}

func (va *Struct) validateRequiredOrSetDefaultIfNeeds(f *jsonflags.StructFields, seen map[string]struct{}, dec *jsontext.Decoder, options jsontext.Options) error {
	errs := make([]error, 0, f.Len())
	prefix := dec.StackPointer()

	for sf := range f.StructField() {
		if _, ok := seen[sf.Name]; ok {
			continue
		}

		validator, err := va.getFieldValidator(sf)
		if err != nil {
			return err
		}

		value := []byte("null")
		if defaultValuer, ok := validator.(WithDefaultValue); ok {
			if defaultValue := defaultValuer.DefaultValue(); len(defaultValue) > 0 {
				value = defaultValue
			} else {
				if optional, ok := validator.(WithOptional); ok {
					if optional.Optional() {
						// skip optional
						continue
					}
				}
			}
		}

		v := &Value{Value: sf.GetOrNewAt(va.Value), Validator: validator}

		jsonOptions := va.patchJsonOptions(sf, options)

		subDec := jsontext.NewDecoder(bytes.NewBuffer(value), jsonOptions)
		if err := v.UnmarshalDecode(subDec, jsonOptions); err != nil {
			pointer := prefix
			if pointer == "" {
				pointer += "/"
			}

			if pointer == "/" {
				pointer += jsontext.Pointer(sf.Name)
			} else {
				pointer += "/" + jsontext.Pointer(sf.Name)
			}

			errs = append(errs, validatorerrors.PrefixJSONPointer(err, pointer))
		}
	}

	return validatorerrors.Join(errs...)
}

func (va *Struct) getFieldValidator(sf *jsonflags.StructField) (Validator, error) {
	return NewWithStructField(sf)
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

	Validator Validator
}

func (va *Array) SetLen(n int) {
	if va.Kind() == reflect.Slice {
		va.Value.SetLen(n)
	}
}

func (va *Array) UnmarshalDecode(dec *jsontext.Decoder, options json.Options) error {
	tok, err := dec.ReadToken()
	if err != nil {
		return err
	}
	k := tok.Kind()
	switch k {
	case 'n':
		if validator := va.Validator; validator != nil {
			if err := validator.Validate(jsontext.Value("null")); err != nil {
				return validatorerrors.PrefixJSONPointer(err, dec.StackPointer())
			}
		}

		va.SetZero()

		return nil
	case '[':
		mustZero := true
		sliceCap := va.Cap() // array
		if sliceCap > 0 {
			va.SetLen(sliceCap)
		}

		var errs []error

		var elemValidator Validator
		if e, ok := va.Validator.(WithElem); ok {
			elemRule := e.Elem()
			elemRule.Type = va.Type().Elem()

			v, err := New(elemRule)
			if err != nil {
				return &json.SemanticError{
					Err:    err,
					GoType: va.Type(),
				}
			}
			elemValidator = v
		}

		i := 0
		for dec.PeekKind() != ']' {
			if i == sliceCap {
				va.Grow(1)
				sliceCap = va.Cap()
				va.SetLen(sliceCap)
				mustZero = false
			}

			itemValue := &Value{Value: va.Index(i), Validator: elemValidator}
			i++
			if mustZero {
				itemValue.SetZero()
			}

			if err := itemValue.UnmarshalDecode(dec, options); err != nil {
				if !validatorerrors.IsValidationError(err) {
					va.SetLen(i)

					return err
				}
				errs = append(errs, err)
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

		if post, ok := va.Validator.(PostValidator); ok {
			if err := post.PostValidate(va.Value); err != nil {
				errs = append(errs, validatorerrors.PrefixJSONPointer(err, dec.StackPointer()))
			}
		}

		return validatorerrors.Join(errs...)
	}

	return &json.SemanticError{JSONKind: k, GoType: va.Type()}
}
