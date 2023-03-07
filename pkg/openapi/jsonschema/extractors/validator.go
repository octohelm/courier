package extractors

import (
	"context"
	"reflect"

	"github.com/octohelm/x/types"

	"github.com/octohelm/courier/pkg/openapi/jsonschema"
	"github.com/octohelm/courier/pkg/ptr"
	"github.com/octohelm/courier/pkg/validator"
)

func BindSchemaValidationByValidateBytes(s *jsonschema.Schema, typ reflect.Type, validateBytes []byte) error {
	fieldValidator, err := validator.Compile(context.Background(), validateBytes, types.FromRType(typ), nil)
	if err != nil {
		return err
	}
	if fieldValidator != nil {
		BindSchemaValidationByValidator(s, fieldValidator)
		s.AddExtension(jsonschema.XTagValidate, string(validateBytes))
	}
	return nil
}

func BindSchemaValidationByValidator(s *jsonschema.Schema, v validator.Validator) {
	if validatorLoader, ok := v.(*validator.ValidatorLoader); ok {
		v = validatorLoader.Validator
	}

	if s == nil {
		return
	}

	switch vt := v.(type) {
	case *validator.UintValidator:
		if len(vt.Enums) > 0 {
			for _, v := range vt.Enums {
				s.Enum = append(s.Enum, v)
			}
			return
		}

		if vt.ExclusiveMinimum || vt.ExclusiveMaximum {
			if vt.ExclusiveMinimum {
				s.ExclusiveMinimum = ptr.Ptr(float64(vt.Minimum))
			}
			if vt.ExclusiveMaximum {
				s.ExclusiveMaximum = ptr.Ptr(float64(vt.Maximum))
			}
		} else {
			s.Minimum = ptr.Ptr(float64(vt.Minimum))
			s.Maximum = ptr.Ptr(float64(vt.Maximum))
		}

		if vt.MultipleOf > 0 {
			s.MultipleOf = ptr.Ptr(float64(vt.MultipleOf))
		}
	case *validator.IntValidator:
		if len(vt.Enums) > 0 {
			for _, v := range vt.Enums {
				s.Enum = append(s.Enum, v)
			}
			return
		}

		if vt.Minimum != nil {
			if vt.ExclusiveMinimum {
				s.ExclusiveMinimum = ptr.Ptr(float64(*vt.Minimum))
			} else {
				s.Minimum = ptr.Ptr(float64(*vt.Minimum))
			}
		}

		if vt.Maximum != nil {
			if vt.ExclusiveMaximum {
				s.ExclusiveMaximum = ptr.Ptr(float64(*vt.Maximum))
			} else {
				s.Maximum = ptr.Ptr(float64(*vt.Maximum))
			}
		}

		if vt.MultipleOf > 0 {
			s.MultipleOf = ptr.Ptr(float64(vt.MultipleOf))
		}
	case *validator.FloatValidator:
		if len(vt.Enums) > 0 {
			for _, v := range vt.Enums {
				s.Enum = append(s.Enum, v)
			}
			return
		}

		if vt.Minimum != nil {
			if vt.ExclusiveMinimum {
				s.ExclusiveMinimum = ptr.Ptr(*vt.Minimum)
			} else {
				s.Minimum = ptr.Ptr(*vt.Minimum)
			}
		}

		if vt.Maximum != nil {
			if vt.ExclusiveMaximum {
				s.ExclusiveMaximum = ptr.Ptr(*vt.Maximum)
			} else {
				s.Maximum = ptr.Ptr(*vt.Maximum)
			}
		}

		if vt.MultipleOf > 0 {
			s.MultipleOf = ptr.Ptr(vt.MultipleOf)
		}
	case *validator.StrfmtValidator:
		s.Type = []string{"string"} // force to type string for TextMarshaler
		s.Format = vt.Names()[0]
	case *validator.StringValidator:
		s.Type = []string{"string"} // force to type string for TextMarshaler

		if len(vt.Enums) > 0 {
			for _, v := range vt.Enums {
				s.Enum = append(s.Enum, v)
			}
			return
		}

		s.MinLength = ptr.Ptr(vt.MinLength)
		if vt.MaxLength != nil {
			s.MaxLength = ptr.Ptr(*vt.MaxLength)
		}
		if vt.Pattern != "" {
			s.Pattern = vt.Pattern
		}
	case *validator.SliceValidator:
		s.MinItems = ptr.Ptr(vt.MinItems)
		if vt.MaxItems != nil {
			s.MaxItems = ptr.Ptr(*vt.MaxItems)
		}

		if vt.ElemValidator != nil {
			if s.Items == nil {
				s.Items = &jsonschema.SchemaOrArray{}
			}

			BindSchemaValidationByValidator(s.Items.Schema, vt.ElemValidator)
		}
	case *validator.MapValidator:
		s.MinProperties = ptr.Ptr(vt.MinProperties)
		if vt.MaxProperties != nil {
			s.MaxProperties = ptr.Ptr(*vt.MaxProperties)
		}
		if vt.ElemValidator != nil {
			BindSchemaValidationByValidator(s.AdditionalProperties.Schema, vt.ElemValidator)
		}
	}
}
