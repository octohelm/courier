package extractors

import (
	"reflect"
	"regexp"
	"testing"

	"github.com/octohelm/x/ptr"

	"github.com/octohelm/courier/pkg/openapi/jsonschema"
	"github.com/octohelm/courier/pkg/validator"
	"github.com/octohelm/courier/pkg/validator/validators"
)

func TestPatchSchemaValidationByValidatorContracts(t *testing.T) {
	t.Run("numeric validators patch number schema", func(t *testing.T) {
		schema := PatchSchemaValidationByValidator(
			&jsonschema.NumberType{Type: "integer"},
			&validators.IntegerValidator[int64]{
				Number: validators.Number[int64]{
					Minimum:          ptr.Ptr(int64(1)),
					Maximum:          ptr.Ptr(int64(10)),
					ExclusiveMaximum: true,
					MultipleOf:       2,
				},
			},
		)

		number, ok := schema.(*jsonschema.NumberType)
		if !ok {
			t.Fatalf("expected number schema, got %T", schema)
		}
		if number.Minimum == nil || *number.Minimum != 1 {
			t.Fatalf("unexpected minimum: %#v", number.Minimum)
		}
		if number.ExclusiveMaximum == nil || *number.ExclusiveMaximum != 10 {
			t.Fatalf("unexpected exclusive maximum: %#v", number.ExclusiveMaximum)
		}
		if number.MultipleOf == nil || *number.MultipleOf != 2 {
			t.Fatalf("unexpected multipleOf: %#v", number.MultipleOf)
		}

		enumSchema := PatchSchemaValidationByValidator(
			&jsonschema.NumberType{Type: "number"},
			&validators.FloatValidator{
				Number: validators.Number[float64]{
					Enums: []float64{1.5, 2.5},
				},
			},
		)
		if enum, ok := enumSchema.(*jsonschema.EnumType); !ok || !reflect.DeepEqual(enum.Enum, []any{1.5, 2.5}) {
			t.Fatalf("unexpected float enum schema: %#v", enumSchema)
		}
	})

	t.Run("string validators patch enum and pattern metadata", func(t *testing.T) {
		enumSchema := PatchSchemaValidationByValidator(
			jsonschema.String(),
			&validators.StringValidator{Enums: []string{"foo", "bar"}},
		)
		if enum, ok := enumSchema.(*jsonschema.EnumType); !ok || !reflect.DeepEqual(enum.Enum, []any{"foo", "bar"}) {
			t.Fatalf("unexpected string enum schema: %#v", enumSchema)
		}

		maxLength := uint64(8)
		patternSchema := PatchSchemaValidationByValidator(
			jsonschema.String(),
			&validators.StringValidator{
				Format:        "uuid",
				MinLength:     2,
				MaxLength:     &maxLength,
				Pattern:       regexp.MustCompile("foo|bar"),
				PatternErrMsg: "必须匹配 foo 或 bar",
			},
		)
		str, ok := patternSchema.(*jsonschema.StringType)
		if !ok {
			t.Fatalf("expected string schema, got %T", patternSchema)
		}
		if str.Format != "uuid" || str.MinLength == nil || *str.MinLength != 2 || str.MaxLength == nil || *str.MaxLength != 8 {
			t.Fatalf("unexpected string validator result: %#v", str)
		}
		if str.Pattern != "foo|bar" {
			t.Fatalf("unexpected pattern: %q", str.Pattern)
		}
		if errMsg, ok := str.GetExtension(jsonschema.XPatternErrMsg); !ok || errMsg != "必须匹配 foo 或 bar" {
			t.Fatalf("unexpected pattern error message extension: %#v %v", errMsg, ok)
		}
	})
}

func TestPatchSchemaValidationContracts(t *testing.T) {
	t.Run("PatchSchemaValidation adds outer validate tag and patches slice items", func(t *testing.T) {
		schema, err := PatchSchemaValidation(
			jsonschema.ArrayOf(jsonschema.String()),
			validator.Option{
				Type: reflect.TypeFor[[]string](),
				Rule: "@slice<@string[1,3]>[1,2]",
			},
		)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		array, ok := schema.(*jsonschema.ArrayType)
		if !ok {
			t.Fatalf("expected array schema, got %T", schema)
		}
		if array.MinItems == nil || *array.MinItems != 1 || array.MaxItems == nil || *array.MaxItems != 2 {
			t.Fatalf("unexpected array bounds: min=%v max=%v", array.MinItems, array.MaxItems)
		}
		if rule, ok := array.GetExtension(jsonschema.XTagValidate); !ok || rule != "@slice<@string[1,3]>[1,2]" {
			t.Fatalf("unexpected x-tag-validate: %#v %v", rule, ok)
		}

		itemSchema, ok := array.Items.(*jsonschema.StringType)
		if !ok {
			t.Fatalf("expected string item schema, got %T", array.Items)
		}
		if itemSchema.MinLength == nil || *itemSchema.MinLength != 1 || itemSchema.MaxLength == nil || *itemSchema.MaxLength != 3 {
			t.Fatalf("unexpected item schema bounds: %#v", itemSchema)
		}
	})

	t.Run("PatchSchemaValidation patches map key and value schema", func(t *testing.T) {
		schema, err := PatchSchemaValidation(
			jsonschema.RecordOf(jsonschema.String(), jsonschema.String()),
			validator.Option{
				Type: reflect.TypeFor[map[string]string](),
				Rule: "@map<@string{foo,bar},@string[1,2]>[1,3]",
			},
		)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		record, ok := schema.(*jsonschema.ObjectType)
		if !ok {
			t.Fatalf("expected object schema, got %T", schema)
		}
		if record.MinProperties == nil || *record.MinProperties != 1 || record.MaxProperties == nil || *record.MaxProperties != 3 {
			t.Fatalf("unexpected property bounds: min=%v max=%v", record.MinProperties, record.MaxProperties)
		}

		keyEnum, ok := record.PropertyNames.(*jsonschema.EnumType)
		if !ok || !reflect.DeepEqual(keyEnum.Enum, []any{"foo", "bar"}) {
			t.Fatalf("unexpected key enum schema: %#v", record.PropertyNames)
		}

		valueSchema, ok := record.AdditionalProperties.(*jsonschema.StringType)
		if !ok {
			t.Fatalf("expected string value schema, got %T", record.AdditionalProperties)
		}
		if valueSchema.MinLength == nil || *valueSchema.MinLength != 1 || valueSchema.MaxLength == nil || *valueSchema.MaxLength != 2 {
			t.Fatalf("unexpected map value schema bounds: %#v", valueSchema)
		}
	})
}
