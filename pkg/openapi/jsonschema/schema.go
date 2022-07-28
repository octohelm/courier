package jsonschema

import (
	"encoding/json"
	"fmt"
	"strings"
)

func NewSchema(tpe string, fmt ...string) *Schema {
	return &Schema{
		SchemaBasic: SchemaBasic{
			Type:   StringOrArray{tpe},
			Format: strings.Join(fmt, ""),
		},
	}
}

func Integer() *Schema {
	return NewSchema(TypeInteger, "int32")
}

func Long() *Schema {
	return NewSchema(TypeInteger, "int64")
}

func Float() *Schema {
	return NewSchema(TypeNumber, "float")
}

func Double() *Schema {
	return NewSchema(TypeNumber, "double")
}

func String() *Schema {
	return NewSchema(TypeString, "")
}

func Bytes() *Schema {
	return NewSchema(TypeString, "bytes")
}

func Binary() *Schema {
	return NewSchema(TypeString, "binary")
}

func Boolean() *Schema {
	return NewSchema(TypeBoolean, "")
}

func ItemsOf(items *Schema) *Schema {
	return &Schema{
		SchemaBasic: SchemaBasic{
			Type: StringOrArray{TypeArray},
			Items: &SchemaOrArray{
				Schema: items,
			},
		},
	}
}

type Props map[string]*Schema

func ObjectOf(props Props, required ...string) *Schema {
	return &Schema{
		SchemaBasic: SchemaBasic{
			Type:       StringOrArray{TypeObject},
			Properties: props,
			SchemaValidation: SchemaValidation{
				Required: required,
			},
		},
	}
}

func MapOf(s *Schema) *Schema {
	return KeyValueOf(nil, s)
}

func KeyValueOf(k *Schema, s *Schema) *Schema {
	return &Schema{
		SchemaBasic: SchemaBasic{
			Type: StringOrArray{TypeObject},
			AdditionalProperties: &SchemaOrBool{
				Allows: true,
				Schema: s,
			},
			PropertyNames: k,
		},
	}
}

func AllOf(schemas ...*Schema) *Schema {
	return &Schema{
		SchemaBasic: SchemaBasic{
			AllOf: schemas,
		},
	}
}

func AnyOf(schemas ...*Schema) *Schema {
	return &Schema{
		SchemaBasic: SchemaBasic{
			AnyOf: schemas,
		},
	}
}

func OneOf(schemas ...*Schema) *Schema {
	return &Schema{
		SchemaBasic: SchemaBasic{
			OneOf: schemas,
		},
	}
}

func Not(schema *Schema) *Schema {
	return &Schema{
		SchemaBasic: SchemaBasic{
			Not: schema,
		},
	}
}

type Schema struct {
	SchemaBasic

	Reference
	VendorExtensible
}

func (s Schema) WithValidation(validation *SchemaValidation) *Schema {
	s.Enum = validation.Enum

	if len(s.Type) == 1 {
		switch s.Type[0] {
		case TypeInteger, TypeNumber:
			s.MultipleOf = validation.MultipleOf
			s.Maximum = validation.Maximum
			s.ExclusiveMaximum = validation.ExclusiveMaximum
			s.Minimum = validation.Minimum
			s.ExclusiveMinimum = validation.ExclusiveMinimum
		case TypeString:
			s.MaxLength = validation.MaxLength
			s.MinLength = validation.MinLength
			s.Pattern = validation.Pattern
		case TypeArray:
			s.MaxItems = validation.MaxItems
			s.MinItems = validation.MinItems
			s.UniqueItems = validation.UniqueItems
		case TypeObject:
			s.MaxProperties = validation.MaxProperties
			s.MinProperties = validation.MinProperties
			if len(s.Properties) > 0 {
				s.Required = validation.Required
			}
		}
	}

	return &s
}

func (s *Schema) SetProperty(name string, propSchema *Schema, required bool) {
	if !s.Type.Contains(TypeObject) {
		return
	}
	if s.Properties == nil {
		s.Properties = make(map[string]*Schema)
	}
	s.Properties[name] = propSchema
	if required {
		s.Required = append(s.Required, name)
	}
}

func (s Schema) WithDesc(desc string) *Schema {
	s.Description = desc
	return &s
}

func (s Schema) WithTitle(title string) *Schema {
	s.Title = title
	return &s
}

func (s Schema) WithDiscriminator(discriminator *Discriminator) *Schema {
	s.Discriminator = discriminator
	return &s
}

func (s Schema) MarshalJSON() ([]byte, error) {
	return s.MarshalJSONRefFirst(s.SchemaBasic, s.VendorExtensible)
}

func (s *Schema) UnmarshalJSON(data []byte) error {
	return s.UnmarshalJSONRefFirst(data, &s.SchemaBasic, &s.VendorExtensible)
}

type SchemaValidation struct {
	// numbers
	MultipleOf       *float64 `json:"multipleOf,omitempty"`
	Maximum          *float64 `json:"maximum,omitempty"`
	ExclusiveMaximum bool     `json:"exclusiveMaximum,omitempty"`
	Minimum          *float64 `json:"minimum,omitempty"`
	ExclusiveMinimum bool     `json:"exclusiveMinimum,omitempty"`

	// string
	MaxLength *uint64 `json:"maxLength,omitempty"`
	MinLength *uint64 `json:"minLength,omitempty"`
	Pattern   string  `json:"pattern,omitempty"`

	// array
	MaxItems    *uint64 `json:"maxItems,omitempty"`
	MinItems    *uint64 `json:"minItems,omitempty"`
	UniqueItems bool    `json:"uniqueItems,omitempty"`

	// object
	MaxProperties *uint64  `json:"maxProperties,omitempty"`
	MinProperties *uint64  `json:"minProperties,omitempty"`
	Required      []string `json:"required,omitempty"`

	// any
	Enum []any `json:"enum,omitempty"`
}

type SchemaBasic struct {
	Type   StringOrArray `json:"type,omitempty"`
	Format string        `json:"format,omitempty"`

	Items                *SchemaOrArray     `json:"items,omitempty"`
	Properties           map[string]*Schema `json:"properties,omitempty"`
	AdditionalProperties *SchemaOrBool      `json:"additionalProperties,omitempty"`
	PropertyNames        *Schema            `json:"propertyNames,omitempty"`

	AllOf []*Schema `json:"allOf,omitempty"`
	AnyOf []*Schema `json:"anyOf,omitempty"`
	OneOf []*Schema `json:"oneOf,omitempty"`
	Not   *Schema   `json:"not,omitempty"`

	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
	Default     any    `json:"default,omitempty"`

	Nullable      bool           `json:"nullable,omitempty"`
	Discriminator *Discriminator `json:"discriminator,omitempty"`
	ReadOnly      bool           `json:"readOnly,omitempty"`
	WriteOnly     bool           `json:"writeOnly,omitempty"`
	XML           *XML           `json:"xml,omitempty"`
	ExternalDocs  *ExternalDoc   `json:"external_docs,omitempty"`
	Example       any            `json:"example,omitempty"`
	Deprecated    bool           `json:"deprecated,omitempty"`

	Definitions map[string]*Schema `json:"definitions,omitempty"`

	SchemaValidation
}

type Discriminator struct {
	PropertyName string            `json:"propertyName"`
	Mapping      map[string]string `json:"mapping,omitempty"`
}

type XML struct {
	Name      string `json:"name,omitempty"`
	Namespace string `json:"namespace,omitempty"`
	Prefix    string `json:"prefix,omitempty"`
	Attribute bool   `json:"attribute,omitempty"`
	Wrapped   bool   `json:"wrapped,omitempty"`
}

const (
	TypeInteger = "integer"
	TypeNumber  = "number"
	TypeString  = "string"
	TypeBoolean = "boolean"

	TypeArray  = "array"
	TypeObject = "object"
)

type SchemaOrBool struct {
	Allows bool
	Schema *Schema
}

func (s *SchemaOrBool) UnmarshalJSON(data []byte) error {
	s.Allows = true
	if len(data) > 0 && data[0] == '{' {
		var schema Schema
		if err := json.Unmarshal(data, &schema); err != nil {
			return err
		}
		s.Schema = &schema
	}
	return nil
}

func (s *SchemaOrBool) MarshalJSON() ([]byte, error) {
	if s.Schema != nil {
		return json.Marshal(s.Schema)
	}
	if s.Schema == nil && !s.Allows {
		return []byte("false"), nil
	}
	return []byte("true"), nil
}

type SchemaOrArray struct {
	Schema  *Schema
	Schemas []Schema
}

func (s SchemaOrArray) Len() int {
	if s.Schema != nil {
		return 1
	}
	return len(s.Schemas)
}

func (s *SchemaOrArray) UnmarshalJSON(data []byte) error {
	var nw SchemaOrArray
	var first byte
	if len(data) > 1 {
		first = data[0]
	}
	if first == '{' {
		var sch Schema
		if err := json.Unmarshal(data, &sch); err != nil {
			return err
		}
		nw.Schema = &sch
	}
	if first == '[' {
		if err := json.Unmarshal(data, &nw.Schemas); err != nil {
			return err
		}
	}
	*s = nw
	return nil
}

func (s *SchemaOrArray) MarshalJSON() ([]byte, error) {
	if len(s.Schemas) > 0 {
		return json.Marshal(s.Schemas)
	}
	return json.Marshal(s.Schema)
}

type StringOrArray []string

func (s StringOrArray) Contains(value string) bool {
	for _, str := range s {
		if str == value {
			return true
		}
	}
	return false
}

func (s *StringOrArray) UnmarshalJSON(data []byte) error {
	var first byte
	if len(data) > 1 {
		first = data[0]
	}

	if first == '[' {
		var parsed []string
		if err := json.Unmarshal(data, &parsed); err != nil {
			return err
		}
		*s = parsed
		return nil
	}

	var single any
	if err := json.Unmarshal(data, &single); err != nil {
		return err
	}
	if single == nil {
		return nil
	}
	switch v := single.(type) {
	case string:
		*s = []string{v}
		return nil
	default:
		return fmt.Errorf("only string or array is allowed, not %T", single)
	}
}

func (s StringOrArray) MarshalJSON() ([]byte, error) {
	if len(s) == 1 {
		return json.Marshal([]string(s)[0])
	}
	return json.Marshal([]string(s))
}
