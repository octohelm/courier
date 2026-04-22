package extractors

import (
	"context"
	"encoding"
	"reflect"
	"strings"
	"testing"

	"github.com/octohelm/courier/pkg/openapi/jsonschema"
)

type schemaRegisterStub struct {
	seen    map[string]bool
	schemas map[string]jsonschema.Schema
}

func (s *schemaRegisterStub) Record(typeRef string) bool {
	if s.seen == nil {
		s.seen = map[string]bool{}
	}
	seen := s.seen[typeRef]
	s.seen[typeRef] = true
	return seen
}

func (s *schemaRegisterStub) RegisterSchema(ref string, schema jsonschema.Schema) {
	if s.schemas == nil {
		s.schemas = map[string]jsonschema.Schema{}
	}
	s.schemas[ref] = schema
}

func (s *schemaRegisterStub) RefString(ref string) string {
	return "#/components/schemas/" + strings.ReplaceAll(ref, "/", ".")
}

type (
	hiddenField string
	modeField   string
)

type documentedSchema struct {
	Name   string      `json:"name,omitempty"`
	Mode   modeField   `json:"mode,omitempty"`
	Hidden hiddenField `json:"hidden,omitempty"`
}

func (documentedSchema) SwaggerDoc() map[string]string {
	return map[string]string{
		"Name": "名称. 用于显示\nopenapi:internal",
		"Mode": "模式. One of fast, slow",
	}
}

type manifestSchema struct {
	Kind       string `json:"kind,omitempty"`
	APIVersion string `json:"apiVersion,omitempty"`
}

func (manifestSchema) GetKind() string {
	return "DemoKind"
}

func (manifestSchema) GetAPIVersion() string {
	return "demo/v1"
}

type formatSchema string

func (formatSchema) OpenAPISchemaFormat() string {
	return "int-or-string"
}

type openAPITypeSchema string

func (openAPITypeSchema) OpenAPISchemaType() []string {
	return []string{"boolean"}
}

type directSchema struct{}

func (directSchema) OpenAPISchema() jsonschema.Schema {
	return jsonschema.Binary()
}

type textSchema string

func (textSchema) MarshalText() ([]byte, error) {
	return []byte("text"), nil
}

func (*textSchema) UnmarshalText([]byte) error {
	return nil
}

var (
	_ encoding.TextMarshaler   = textSchema("")
	_ encoding.TextUnmarshaler = (*textSchema)(nil)
)

type unionSchema struct{}

func (unionSchema) OneOf() []any {
	return []any{directSchema{}, openAPITypeSchema("")}
}

type taggedUnionSchema struct{}

func (taggedUnionSchema) Discriminator() string {
	return "kind"
}

func (taggedUnionSchema) Mapping() map[string]any {
	return map[string]any{
		"binary": directSchema{},
		"bool":   openAPITypeSchema(""),
	}
}

func (taggedUnionSchema) SetUnderlying(any) {}

func TestSchemaFromAndFieldFilters(t *testing.T) {
	if got := SchemaFrom(context.Background(), nil, true); got != nil {
		t.Fatalf("expected nil schema, got %T", got)
	}

	RegisterFieldFilter(reflect.TypeFor[*hiddenField](), FieldFilter{Exclude: []string{"Hidden"}})

	if FieldShouldPick(reflect.TypeFor[hiddenField](), "Hidden") {
		t.Fatalf("expected hidden field to be filtered out")
	}
	if !FieldShouldPick(reflect.TypeFor[string](), "Name") {
		t.Fatalf("expected regular string field to be kept")
	}
}

func TestSetTitleOrDescriptionAndPickStringEnum(t *testing.T) {
	SetTitleOrDescription(nil, []string{"ignored"})

	metadata := &jsonschema.Metadata{}
	SetTitleOrDescription(metadata, []string{"标题", "第一行", "openapi:hide", "第二行"})

	if metadata.Title != "标题" {
		t.Fatalf("unexpected title: %q", metadata.Title)
	}
	if metadata.Description != "第一行\n第二行" {
		t.Fatalf("unexpected description: %q", metadata.Description)
	}

	if got := pickStringEnumFromDesc(`模式. One of fast, slow`); !reflect.DeepEqual(got, []string{"fast", "slow"}) {
		t.Fatalf("unexpected enum values: %#v", got)
	}
	if got := pickStringEnumFromDesc(`状态. Can be "ready" or "done"`); !reflect.DeepEqual(got, []string{"ready", "done"}) {
		t.Fatalf("unexpected quoted enum values: %#v", got)
	}
}

func TestSchemaFromTypeBuildsDocumentedStructSchema(t *testing.T) {
	register := &schemaRegisterStub{}
	ctx := SchemaRegisterContext.Inject(context.Background(), register)
	schema := SchemaFromType(ctx, reflect.TypeFor[documentedSchema](), Opt{Decl: true})

	obj, ok := schema.(*jsonschema.ObjectType)
	if !ok {
		t.Fatalf("expected object schema, got %T", schema)
	}

	if _, ok := obj.GetExtension(jsonschema.XGoVendorType); !ok {
		t.Fatalf("expected x-go-vendor-type extension")
	}

	nameSchema, ok := obj.Properties.Get("name")
	if !ok {
		t.Fatalf("expected Name property")
	}
	nameString, ok := nameSchema.(*jsonschema.StringType)
	if !ok {
		t.Fatalf("expected Name to be string schema, got %T", nameSchema)
	}
	if nameString.Title != "名称" {
		t.Fatalf("unexpected Name title: %q", nameString.Title)
	}
	if nameString.Description != "用于显示\nopenapi:internal" {
		t.Fatalf("unexpected Name description: %q", nameString.Description)
	}
	if fieldName, ok := nameString.GetExtension(jsonschema.XGoFieldName); !ok || fieldName != "Name" {
		t.Fatalf("unexpected field name extension: %#v %v", fieldName, ok)
	}

	modeSchema, ok := obj.Properties.Get("mode")
	if !ok {
		t.Fatalf("expected Mode property")
	}
	modeRef, ok := modeSchema.(*jsonschema.RefType)
	if !ok {
		t.Fatalf("expected Mode to be ref schema, got %T", modeSchema)
	}
	if modeRef.Ref == nil {
		t.Fatalf("expected Mode ref uri")
	}

	foundEnum := false
	for _, schema := range register.schemas {
		if enumSchema, ok := schema.(*jsonschema.EnumType); ok {
			if reflect.DeepEqual(enumSchema.Enum, []any{"fast", "slow"}) {
				foundEnum = true
				break
			}
		}
	}
	if !foundEnum {
		t.Fatalf("expected registered enum schema for Mode")
	}

	if _, ok := obj.Properties.Get("hidden"); ok {
		t.Fatalf("expected Hidden property to be filtered out")
	}
}

func TestSchemaFromTypeCoversKindsAndSpecialInterfaces(t *testing.T) {
	t.Run("primitive and container kinds", func(t *testing.T) {
		cases := []struct {
			name  string
			typ   reflect.Type
			check func(t *testing.T, s jsonschema.Schema)
		}{
			{
				name: "interface",
				typ:  reflect.TypeFor[any](),
				check: func(t *testing.T, s jsonschema.Schema) {
					if _, ok := s.(*jsonschema.AnyType); !ok {
						t.Fatalf("expected any schema, got %T", s)
					}
				},
			},
			{
				name: "bool",
				typ:  reflect.TypeFor[bool](),
				check: func(t *testing.T, s jsonschema.Schema) {
					if _, ok := s.(*jsonschema.BooleanType); !ok {
						t.Fatalf("expected boolean schema, got %T", s)
					}
				},
			},
			{
				name: "float32",
				typ:  reflect.TypeFor[float32](),
				check: func(t *testing.T, s jsonschema.Schema) {
					number, ok := s.(*jsonschema.NumberType)
					if !ok {
						t.Fatalf("expected number schema, got %T", s)
					}
					if format, ok := number.GetExtension("x-format"); !ok || format != "float32" {
						t.Fatalf("unexpected float format: %#v %v", format, ok)
					}
				},
			},
			{
				name: "int64",
				typ:  reflect.TypeFor[int64](),
				check: func(t *testing.T, s jsonschema.Schema) {
					number, ok := s.(*jsonschema.NumberType)
					if !ok {
						t.Fatalf("expected integer schema, got %T", s)
					}
					if format, ok := number.GetExtension("x-format"); !ok || format != "int64" {
						t.Fatalf("unexpected integer format: %#v %v", format, ok)
					}
				},
			},
			{
				name: "fixed array",
				typ:  reflect.TypeFor[[2]string](),
				check: func(t *testing.T, s jsonschema.Schema) {
					array, ok := s.(*jsonschema.ArrayType)
					if !ok {
						t.Fatalf("expected array schema, got %T", s)
					}
					if array.MinItems == nil || *array.MinItems != 2 || array.MaxItems == nil || *array.MaxItems != 2 {
						t.Fatalf("unexpected fixed array bounds: min=%v max=%v", array.MinItems, array.MaxItems)
					}
				},
			},
			{
				name: "bytes slice",
				typ:  reflect.TypeFor[[]byte](),
				check: func(t *testing.T, s jsonschema.Schema) {
					stringSchema, ok := s.(*jsonschema.StringType)
					if !ok {
						t.Fatalf("expected bytes schema, got %T", s)
					}
					if stringSchema.Format != "bytes" {
						t.Fatalf("unexpected bytes format: %q", stringSchema.Format)
					}
				},
			},
			{
				name: "string slice",
				typ:  reflect.TypeFor[[]string](),
				check: func(t *testing.T, s jsonschema.Schema) {
					array, ok := s.(*jsonschema.ArrayType)
					if !ok {
						t.Fatalf("expected array schema, got %T", s)
					}
					if _, ok := array.Items.(*jsonschema.StringType); !ok {
						t.Fatalf("expected string items, got %T", array.Items)
					}
				},
			},
			{
				name: "string map",
				typ:  reflect.TypeFor[map[string]int](),
				check: func(t *testing.T, s jsonschema.Schema) {
					record, ok := s.(*jsonschema.ObjectType)
					if !ok {
						t.Fatalf("expected record schema, got %T", s)
					}
					if _, ok := record.PropertyNames.(*jsonschema.StringType); !ok {
						t.Fatalf("expected string property names, got %T", record.PropertyNames)
					}
					if number, ok := record.AdditionalProperties.(*jsonschema.NumberType); !ok || number.Type != "integer" {
						t.Fatalf("unexpected additional properties schema: %T", record.AdditionalProperties)
					}
				},
			},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				tc.check(t, SchemaFromType(context.Background(), tc.typ, Opt{Decl: true}))
			})
		}
	})

	t.Run("manifest kind and apiVersion are narrowed to enum", func(t *testing.T) {
		schema := SchemaFromType(context.Background(), reflect.TypeFor[manifestSchema](), Opt{Decl: true})
		obj := schema.(*jsonschema.ObjectType)

		kindSchema, ok := obj.Properties.Get("kind")
		if !ok {
			t.Fatalf("expected Kind property")
		}
		if enumSchema, ok := kindSchema.(*jsonschema.EnumType); !ok || !reflect.DeepEqual(enumSchema.Enum, []any{"DemoKind"}) {
			t.Fatalf("unexpected Kind schema: %#v", kindSchema)
		}

		apiVersionSchema, ok := obj.Properties.Get("apiVersion")
		if !ok {
			t.Fatalf("expected APIVersion property")
		}
		if enumSchema, ok := apiVersionSchema.(*jsonschema.EnumType); !ok || !reflect.DeepEqual(enumSchema.Enum, []any{"demo/v1"}) {
			t.Fatalf("unexpected APIVersion schema: %#v", apiVersionSchema)
		}
	})

	t.Run("named type declaration false becomes ref and records schema", func(t *testing.T) {
		register := &schemaRegisterStub{}
		ctx := SchemaRegisterContext.Inject(context.Background(), register)

		schema := SchemaFromType(ctx, reflect.TypeFor[**documentedSchema](), Opt{})
		ref, ok := schema.(*jsonschema.RefType)
		if !ok {
			t.Fatalf("expected ref schema, got %T", schema)
		}
		if ref.Ref == nil {
			t.Fatalf("expected ref uri")
		}
		if starLevel, ok := ref.GetExtension(jsonschema.XGoStarLevel); !ok || starLevel != 2 {
			t.Fatalf("unexpected pointer star level: %#v %v", starLevel, ok)
		}
		foundObject := false
		for _, schema := range register.schemas {
			if _, ok := schema.(*jsonschema.ObjectType); ok {
				foundObject = true
				break
			}
		}
		if !foundObject {
			t.Fatalf("expected registered object schema")
		}
	})

	t.Run("special schema interfaces are honored", func(t *testing.T) {
		if union, ok := SchemaFromType(context.Background(), reflect.TypeFor[formatSchema](), Opt{Decl: true}).(*jsonschema.UnionType); !ok || len(union.OneOf) != 2 {
			t.Fatalf("expected int-or-string union schema")
		}

		if _, ok := SchemaFromType(context.Background(), reflect.TypeFor[openAPITypeSchema](), Opt{Decl: true}).(*jsonschema.BooleanType); !ok {
			t.Fatalf("expected boolean schema from OpenAPISchemaType")
		}

		if binary, ok := SchemaFromType(context.Background(), reflect.TypeFor[directSchema](), Opt{Decl: true}).(*jsonschema.StringType); !ok || binary.Format != "binary" {
			t.Fatalf("expected binary schema from OpenAPISchema")
		}

		if _, ok := SchemaFromType(context.Background(), reflect.TypeFor[textSchema](), Opt{Decl: true}).(*jsonschema.StringType); !ok {
			t.Fatalf("expected string schema for text marshaler/unmarshaler")
		}

		if union, ok := SchemaFromType(context.Background(), reflect.TypeFor[unionSchema](), Opt{Decl: true}).(*jsonschema.UnionType); !ok || len(union.OneOf) != 2 {
			t.Fatalf("expected oneOf union schema")
		}

		tagged, ok := SchemaFromType(context.Background(), reflect.TypeFor[taggedUnionSchema](), Opt{Decl: true}).(*jsonschema.UnionType)
		if !ok {
			t.Fatalf("expected tagged union schema")
		}
		if tagged.Discriminator == nil || tagged.Discriminator.PropertyName != "kind" {
			t.Fatalf("unexpected tagged union discriminator: %#v", tagged.Discriminator)
		}
		if len(tagged.Discriminator.Mapping) != 2 {
			t.Fatalf("unexpected tagged union mapping: %#v", tagged.Discriminator.Mapping)
		}
	})
}
