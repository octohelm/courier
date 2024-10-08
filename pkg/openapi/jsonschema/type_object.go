package jsonschema

import (
	"io"
	"sort"

	"github.com/go-json-experiment/json"
	"github.com/go-json-experiment/json/jsontext"
)

func ObjectOf(props map[string]Schema, required ...string) *ObjectType {
	return &ObjectType{
		Type:       "object",
		Properties: props,
		Required:   required,
	}
}

func RecordOf(k Schema, v Schema) *ObjectType {
	return &ObjectType{
		Type:                 "object",
		PropertyNames:        k,
		AdditionalProperties: v,
	}
}

func MapOf(v Schema) *ObjectType {
	return RecordOf(nil, v)
}

type ObjectType struct {
	Type string `json:"type,omitzero"`

	Properties           Props  `json:"properties,omitzero"`
	PropertyNames        Schema `json:"propertyNames,omitzero"`
	AdditionalProperties Schema `json:"additionalProperties,omitzero"`

	// validate
	Required      []string `json:"required,omitzero"`
	MaxProperties *uint64  `json:"maxProperties,omitzero"`
	MinProperties *uint64  `json:"minProperties,omitzero"`

	Core
	Metadata
}

type Props map[string]Schema

func (props Props) MarshalJSONV2(encoder *jsontext.Encoder, options json.Options) error {
	keys := make([]string, 0, len(props))
	for k := range props {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	if err := encoder.WriteToken(jsontext.ObjectStart); err != nil {
		return err
	}

	for _, k := range keys {
		if err := json.MarshalEncode(encoder, k); err != nil {
			return err
		}

		if err := json.MarshalEncode(encoder, props[k]); err != nil {
			return err
		}
	}

	if err := encoder.WriteToken(jsontext.ObjectEnd); err != nil {
		return err
	}

	return nil
}

var _ json.MarshalerV2 = Props{}

func (t ObjectType) PrintTo(w io.Writer, optionFns ...SchemaPrintOption) {
	opt := printOption{}
	opt.Build(optionFns...)

	Print(w, func(p Printer) {
		if dynamicAnchor := t.DynamicAnchor; dynamicAnchor != "" {
			p.Printf("#%s: ", dynamicAnchor)
		} else if t.ID != nil {
			p.Printf("#%s: ", t.ID.RefName())
		}

		p.Print("{")
		p.Return()

		defer func() {
			p.Return()
			p.Print("}")
		}()

		end := p.Indent()
		defer end()

		propIdx := 0

		propNames := make([]string, 0, len(t.Properties))
		for name := range t.Properties {
			propNames = append(propNames, name)
		}
		sort.Strings(propNames)

		for propIdx, propName := range propNames {
			prop := t.Properties[propName]
			if propIdx > 0 {
				p.Return()
			}

			if opt.WithDoc {
				if desc := prop.GetMetadata().Description; desc != "" {
					p.PrintDoc(desc)
				}
			}

			p.Printf("%q", propName)

			required := false
			for _, r := range t.Required {
				if r == propName {
					required = true
					break
				}
			}

			if !required {
				p.Print("?")
			}

			p.Print(": ")
			p.PrintFrom(prop, optionFns...)
		}

		if additionalProperties := t.AdditionalProperties; additionalProperties != nil {
			if len(propNames) > 0 {
				p.Return()
			}

			propSchema := t.PropertyNames
			if propSchema == nil {
				propSchema = &StringType{}
			}

			p.Print("[X=")
			p.PrintFrom(propSchema, optionFns...)
			p.Print("]: ")
			p.PrintFrom(additionalProperties, optionFns...)
		}

		for name, d := range t.Defs {
			if propIdx > 0 {
				p.Return()
			}
			if dynamicAnchor := d.GetCore().DynamicAnchor; dynamicAnchor != "" {
				p.Printf("#%s: ", dynamicAnchor)
			} else {
				p.Printf("#%s: ", name)
			}
			p.PrintFrom(d, optionFns...)
			propIdx++
		}
	})
}

func (t *ObjectType) SetProperty(name string, schema Schema, required bool) {
	if t.Properties == nil {
		t.Properties = map[string]Schema{}
	}

	t.Properties[name] = schema

	if required {
		t.Required = append(t.Required, name)
	}
}
