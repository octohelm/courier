package extractors

import (
	"context"
	"reflect"

	"github.com/octohelm/x/types"

	"github.com/octohelm/courier/pkg/openapi/jsonschema"
	"github.com/octohelm/courier/pkg/ptr"
	"github.com/octohelm/courier/pkg/validator"
)

func PatchSchemaValidationByValidateBytes(s jsonschema.Schema, typ reflect.Type, validateBytes []byte) (jsonschema.Schema, error) {
	fieldValidator, err := validator.Compile(context.Background(), validateBytes, types.FromRType(typ), nil)
	if err != nil {
		return nil, err
	}
	if fieldValidator != nil && s != nil {
		s = PatchSchemaValidationByValidator(s, fieldValidator)
		s.GetMetadata().AddExtension(jsonschema.XTagValidate, string(validateBytes))
	}
	return s, nil
}

func PatchSchemaValidationByValidator(s jsonschema.Schema, v validator.Validator) jsonschema.Schema {
	if validatorLoader, ok := v.(*validator.ValidatorLoader); ok {
		v = validatorLoader.Validator
	}

	switch vt := v.(type) {
	case *validator.UintValidator:
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
			if vt.ExclusiveMinimum || vt.ExclusiveMaximum {
				if vt.ExclusiveMinimum {
					x.ExclusiveMinimum = ptr.Ptr(float64(vt.Minimum))
				}
				if vt.ExclusiveMaximum {
					x.ExclusiveMaximum = ptr.Ptr(float64(vt.Maximum))
				}
			} else {
				x.Minimum = ptr.Ptr(float64(vt.Minimum))
				x.Maximum = ptr.Ptr(float64(vt.Maximum))
			}

			if vt.MultipleOf > 0 {
				x.MultipleOf = ptr.Ptr(float64(vt.MultipleOf))
			}
			return x
		}
	case *validator.IntValidator:
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
				if vt.ExclusiveMinimum {
					x.ExclusiveMinimum = ptr.Ptr(float64(*vt.Minimum))
				} else {
					x.Minimum = ptr.Ptr(float64(*vt.Minimum))
				}
			}

			if vt.Maximum != nil {
				if vt.ExclusiveMaximum {
					x.ExclusiveMaximum = ptr.Ptr(float64(*vt.Maximum))
				} else {
					x.Maximum = ptr.Ptr(float64(*vt.Maximum))
				}
			}

			if vt.MultipleOf > 0 {
				x.MultipleOf = ptr.Ptr(float64(vt.MultipleOf))
			}

			return x
		}
	case *validator.FloatValidator:
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
				if vt.ExclusiveMinimum {
					x.ExclusiveMinimum = ptr.Ptr(float64(*vt.Minimum))
				} else {
					x.Minimum = ptr.Ptr(float64(*vt.Minimum))
				}
			}

			if vt.Maximum != nil {
				if vt.ExclusiveMaximum {
					x.ExclusiveMaximum = ptr.Ptr(float64(*vt.Maximum))
				} else {
					x.Maximum = ptr.Ptr(float64(*vt.Maximum))
				}
			}

			if vt.MultipleOf > 0 {
				x.MultipleOf = ptr.Ptr(float64(vt.MultipleOf))
			}

			return x
		}
	case *validator.StrfmtValidator:
		// force to type string for TextMarshaler
		return &jsonschema.StringType{
			Type:   "string",
			Format: vt.Names()[0],
		}
	case *validator.StringValidator:
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
	case *validator.SliceValidator:
		switch x := s.(type) {
		case *jsonschema.ArrayType:

			x.MinItems = ptr.Ptr(vt.MinItems)
			if vt.MaxItems != nil {
				x.MaxItems = ptr.Ptr(*vt.MaxItems)
			}

			if vt.ElemValidator != nil {
				x.Items = PatchSchemaValidationByValidator(x.Items, vt.ElemValidator)
			}

			return x
		}
	case *validator.MapValidator:
		switch x := s.(type) {
		case *jsonschema.ObjectType:
			x.MinProperties = ptr.Ptr(vt.MinProperties)
			if vt.MaxProperties != nil {
				x.MaxProperties = ptr.Ptr(*vt.MaxProperties)
			}
			if vt.ElemValidator != nil {
				x.AdditionalProperties = PatchSchemaValidationByValidator(x.AdditionalProperties, vt.ElemValidator)
			}
			return x
		}
	}

	return s
}
