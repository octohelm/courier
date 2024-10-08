package extractors

import (
	"github.com/octohelm/courier/pkg/openapi/jsonschema"
	"github.com/octohelm/courier/pkg/validator"
	"github.com/octohelm/courier/pkg/validator/validators"
	"github.com/octohelm/x/ptr"
)

func PatchSchemaValidation(s jsonschema.Schema, opt validator.Option) (jsonschema.Schema, error) {
	fieldValidator, err := validator.New(opt)
	if err != nil {
		return nil, err
	}
	if opt.Rule != "" {
		s = PatchSchemaValidationByValidator(s, fieldValidator)
		s.GetMetadata().AddExtension(jsonschema.XTagValidate, opt.Rule)
	}
	return s, nil
}

func PatchSchemaValidationByValidator(s jsonschema.Schema, v validator.Validator) jsonschema.Schema {
	if u, ok := v.(interface{ Unwrap() validator.Validator }); ok {
		return PatchSchemaValidationByValidator(s, u.Unwrap())
	}

	switch vt := v.(type) {
	case *validators.IntegerValidator[int64]:
		return patch(s, vt.Number)
	case *validators.IntegerValidator[uint64]:
		return patch(s, vt.Number)
	case *validators.FloatValidator:
		return patch(s, vt.Number)
	case *validators.StrfmtValidator:
		return &jsonschema.StringType{
			Type:   "string",
			Format: vt.Format(),
		}
	case *validators.StringValidator:
		if len(vt.Enums) > 0 {
			enum := &jsonschema.EnumType{
				Enum: make([]any, len(vt.Enums)),
			}
			for i, v := range vt.Enums {
				enum.Enum[i] = v
			}
			return enum
		}

		s := jsonschema.String()

		s.MinLength = ptr.Ptr(vt.MinLength)

		if vt.MaxLength != nil {
			s.MaxLength = ptr.Ptr(*vt.MaxLength)
		}

		if vt.Pattern != "" {
			s.Pattern = vt.Pattern
		}
		return s
	case *validators.SliceValidator:
		switch x := s.(type) {
		case *jsonschema.ArrayType:

			x.MinItems = ptr.Ptr(vt.MinItems)
			if vt.MaxItems != nil {
				x.MaxItems = ptr.Ptr(*vt.MaxItems)
			}

			elem, err := PatchSchemaValidation(x.Items, vt.Elem())
			if err != nil {
				panic(err)
			}
			x.Items = elem

			return x
		}
	case *validators.MapValidator:
		switch x := s.(type) {
		case *jsonschema.ObjectType:
			x.MinProperties = ptr.Ptr(vt.MinProperties)
			if vt.MaxProperties != nil {
				x.MaxProperties = ptr.Ptr(*vt.MaxProperties)
			}

			key, err := PatchSchemaValidation(x.PropertyNames, vt.Key())
			if err != nil {
				panic(err)
			}
			x.PropertyNames = key

			elem, err := PatchSchemaValidation(x.AdditionalProperties, vt.Key())
			if err != nil {
				panic(err)
			}
			x.AdditionalProperties = elem

			return x
		}
	}

	return s
}

func patch[T ~int64 | ~uint64 | ~float64](s jsonschema.Schema, vt validators.Number[T]) jsonschema.Schema {
	if len(vt.Enums) > 0 {
		enum := &jsonschema.EnumType{
			Enum: make([]any, len(vt.Enums)),
		}
		for i, v := range vt.Enums {
			enum.Enum[i] = v
		}
		return enum
	}

	switch x := s.(type) {
	case *jsonschema.NumberType:
		if vt.Minimum != nil {
			m := *vt.Minimum
			if vt.ExclusiveMinimum {
				x.ExclusiveMinimum = ptr.Ptr(float64(m))
			} else {
				x.Minimum = ptr.Ptr(float64(m))
			}
		}

		if vt.Maximum != nil {
			m := *vt.Maximum
			if vt.ExclusiveMaximum {
				x.ExclusiveMaximum = ptr.Ptr(float64(m))
			} else {
				x.Maximum = ptr.Ptr(float64(m))
			}
		}

		if vt.MultipleOf > 0 {
			x.MultipleOf = ptr.Ptr(float64(vt.MultipleOf))
		}
		return x
	}

	return s
}
