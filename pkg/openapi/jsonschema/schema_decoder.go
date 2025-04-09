package jsonschema

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/go-json-experiment/json"
	"github.com/go-json-experiment/json/jsontext"
)

var (
	ErrInvalidJSONSchemaObject = errors.New("invalid json schema object")
	ErrInvalidJSONSchemaType   = errors.New("invalid json schema type")
)

var schemaUnmarshalers = json.UnmarshalFromFunc[*Schema](func(decoder *jsontext.Decoder, schema *Schema) error {
	return (&schemaDecoder{schema: schema}).UnmarshalJSONFrom(decoder)
})

type schemaDecoder struct {
	schema  *Schema
	options json.Options
	anchors map[string]string
}

var _ json.UnmarshalerFrom = &schemaDecoder{}

func (u *schemaDecoder) UnmarshalJSONFrom(decoder *jsontext.Decoder) error {
	u.options = decoder.Options()

	startToken, err := decoder.ReadToken()
	if err != nil {
		return err
	}

	switch startToken.Kind() {
	case 't':
		// true
		*u.schema = &AnyType{}
		return nil
	case '{':
		// object
		return u.unmarshalFromObject(decoder)
	}

	return ErrInvalidJSONSchemaObject
}

func (u *schemaDecoder) decode(decoder *jsontext.Decoder, target any) error {
	k, err := decoder.ReadValue()
	if err != nil {
		return err
	}
	if err := json.Unmarshal(k, target, u.options); err != nil {
		return err
	}
	return nil
}

func (u *schemaDecoder) schemaOfType(typ string, format string) (Schema, error) {
	switch typ {
	case "array":
		return &ArrayType{Type: typ}, nil
	case "object":
		return &ObjectType{Type: typ}, nil
	case "number":
		t := &NumberType{Type: "number"}
		switch format {
		case "int64", "int32", "int16", "int8", "uint64", "uint32", "uint16", "uint8":
			t.AddExtension("x-format", format)
		case "float":
			t.AddExtension("x-format", "float32")
		case "double":
			t.AddExtension("x-format", "float64")
		}
		return t, nil
	case "integer", "int":
		t := &NumberType{Type: "integer"}
		switch format {
		case "int64", "int32", "int16", "int8", "uint64", "uint32", "uint16", "uint8":
			t.AddExtension("x-format", format)
		}
		return t, nil
	case "string":
		return &StringType{Type: typ}, nil
	case "null":
		return &NullType{Type: typ}, nil
	case "boolean":
		return &BooleanType{Type: typ}, nil
	}
	return nil, ErrInvalidJSONSchemaType
}

func (u *schemaDecoder) unmarshalFromObject(decoder *jsontext.Decoder) error {
	unprocessed := bytes.NewBuffer(nil)
	unprocessedEnc := jsontext.NewEncoder(unprocessed)

	_ = unprocessedEnc.WriteToken(jsontext.BeginObject)

	var schema any
	var typ string
	var format string
	var additionalSchemas []Schema

	for kind := decoder.PeekKind(); kind != '}'; kind = decoder.PeekKind() {
		var prop string
		if err := u.decode(decoder, &prop); err != nil {
			return fmt.Errorf("decode prop failed: %w", err)
		}

		// renaming
		switch prop {
		case "$recursiveRef":
			prop = "$dynamicRef"
		case "$recursiveAnchor":
			prop = "$dynamicAnchor"
		case "definitions":
			prop = "$def"
		case "dependencies":
			// TODO convert to with dependentSchemas and dependentRequired
		}

		switch prop {
		case "const":
			var value any
			if err := u.decode(decoder, &value); err != nil {
				return fmt.Errorf("decode prop %s failed: %w", prop, err)
			}
			schema = &EnumType{
				Enum: []any{value},
			}
			// skip unmarshal decode const
			continue
		case "format":
			if err := u.decode(decoder, &format); err != nil {
				return fmt.Errorf("decode prop %s failed: %w", prop, err)
			}
			continue
		case "enum":
			schema = &EnumType{}
		case "items", "prefixItems":
			schema = &ArrayType{
				Type: "array",
			}
		case "properties", "propertyNames", "patternProperties", "additionalProperties", "required":
			schema = &ObjectType{Type: "object"}
		case "oneOf", "discriminator":
			schema = &UnionType{}
		case "allOf":
			schema = &IntersectionType{}
		case "$dynamicRef":
			schema = &RefType{}
		case "$ref":
			schema = &RefType{}
		case "type":
			v, err := decoder.ReadValue()
			if err != nil {
				return err
			}
			switch v.Kind() {
			case '[':
				var unionType []string

				if err := json.Unmarshal(v, &unionType); err != nil {
					return err
				}

				if len(unionType) > 0 {
					typ = unionType[0]
				}

				for i, t := range unionType {
					if i == 0 {
						typ = t
						continue
					}

					s, err := u.schemaOfType(t, "")
					if err != nil {
						return err
					}

					additionalSchemas = append(additionalSchemas, s)
				}

				continue
			default:
				if err := json.Unmarshal(v, &typ); err != nil {
					return err
				}
			}
			continue
		}

		v, err := decoder.ReadValue()
		if err != nil {
			return fmt.Errorf("read prop %s failed: %w", prop, err)
		}

		_ = json.MarshalEncode(unprocessedEnc, prop)
		_ = json.MarshalEncode(unprocessedEnc, v)
	}

	// read the EndObject to mark decoder finished
	t, err := decoder.ReadToken()
	if err != nil {
		return err
	}
	_ = unprocessedEnc.WriteToken(t)

	if schema == nil || len(additionalSchemas) == 0 {
		if typ != "" {
			s, err := u.schemaOfType(typ, format)
			if err != nil {
				return err
			}
			schema = s
		}
	}

	if schema == nil {
		schema = &AnyType{}
	}

	// {}\n
	if unprocessed.Len() > 3 {
		if err := json.UnmarshalRead(unprocessed, schema, u.options); err != nil {
			return err
		}
	}

	if it, ok := schema.(*IntersectionType); ok {
		// TODO for old structure
		if len(it.AllOf) == 2 {
			switch x := it.AllOf[1].(type) {
			case *ObjectType:
				// skip
			default:
				s := it.AllOf[0]
				x.GetMetadata().DeepCopyInto(s.GetMetadata())
				schema = s
			}
		}
	}

	if len(additionalSchemas) > 0 {
		*u.schema = OneOf(
			append([]Schema{
				schema.(Schema),
			}, additionalSchemas...)...,
		)
	} else {
		*u.schema = schema.(Schema)
	}

	return nil
}
