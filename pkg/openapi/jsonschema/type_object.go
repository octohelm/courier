package jsonschema

import (
	"errors"
	"fmt"
	"io"
	"iter"
	"maps"
	"slices"
	"strconv"

	"github.com/go-json-experiment/json"
	"github.com/go-json-experiment/json/jsontext"
	"github.com/octohelm/x/container/list"
)

func ObjectOf(props map[string]Schema, required ...string) *ObjectType {
	properties := Props{}

	for _, prop := range slices.Sorted(maps.Keys(props)) {
		properties.Set(prop, props[prop])
	}

	return &ObjectType{
		Type:       "object",
		Properties: properties,
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

type field struct {
	key   string
	value Schema
}

type Props struct {
	props   map[string]*list.Element[*field]
	ll      list.List[*field]
	created bool
}

func (p Props) IsZero() bool {
	return len(p.props) == 0
}

func (p *Props) Len() int {
	return len(p.props)
}

func (p *Props) KeyValues() iter.Seq2[string, Schema] {
	return func(yield func(string, Schema) bool) {
		for el := p.ll.Front(); el != nil; el = el.Next() {
			if !yield(el.Value.key, el.Value.value) {
				return
			}
		}
	}
}

func (p *Props) Get(key string) (Schema, bool) {
	if p.props != nil {
		v, ok := p.props[key]
		if ok {
			return v.Value.value, true
		}
	}
	return nil, false
}

func (p *Props) initOnce() {
	if !p.created {
		p.created = true

		p.props = map[string]*list.Element[*field]{}
		p.ll.Init()
	}
}

func (p *Props) Set(key string, value Schema) bool {
	p.initOnce()

	_, alreadyExist := p.props[key]
	if alreadyExist {
		p.props[key].Value.value = value
		return false
	}

	element := &field{key: key, value: value}
	p.props[key] = p.ll.PushBack(element)
	return true
}

func (p *Props) Delete(key string) (didDelete bool) {
	if p.props == nil {
		return false
	}

	element, ok := p.props[key]
	if ok {
		p.ll.Remove(element)

		delete(p.props, key)
	}
	return ok
}

var _ json.UnmarshalerFrom = &Props{}

func (props *Props) UnmarshalJSONFrom(d *jsontext.Decoder) error {
	t, err := d.ReadToken()
	if err != nil {
		if err == io.EOF {
			return nil
		}
		return err
	}

	kind := t.Kind()

	if kind != '{' {
		return &json.SemanticError{
			JSONPointer: d.StackPointer(),
			Err:         fmt.Errorf("object should starts with `{`, but got `%s`", kind),
		}
	}

	if props == nil {
		*props = Props{}
	}

	for kind := d.PeekKind(); kind != '}'; kind = d.PeekKind() {
		k, err := d.ReadValue()
		if err != nil {
			return err
		}

		key, err := strconv.Unquote(string(k))
		if err != nil {
			return &json.SemanticError{
				JSONPointer: d.StackPointer(),
				Err:         errors.New("key should be quoted string"),
			}
		}

		var schema Schema
		if err := json.UnmarshalDecode(d, &schema, json.WithUnmarshalers(schemaUnmarshalers)); err != nil {
			return err
		}

		props.Set(key, schema)
	}

	// read the close '}'
	if _, err := d.ReadToken(); err != nil {
		if err != io.EOF {
			return nil
		}
		return err
	}
	return nil
}

var _ json.MarshalerTo = Props{}

func (p Props) MarshalJSONTo(encoder *jsontext.Encoder) error {
	if err := encoder.WriteToken(jsontext.BeginObject); err != nil {
		return err
	}

	for name, s := range p.KeyValues() {
		if err := json.MarshalEncode(encoder, name); err != nil {
			return err
		}

		if err := json.MarshalEncode(encoder, s); err != nil {
			return err
		}
	}

	if err := encoder.WriteToken(jsontext.EndObject); err != nil {
		return err
	}

	return nil
}

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
		for propName, prop := range t.Properties.KeyValues() {
			if propIdx > 0 {
				p.Return()
			}

			propIdx++

			if opt.WithDoc {
				if title := prop.GetMetadata().Title; title != "" {
					p.PrintDoc(title)
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
			if !t.Properties.IsZero() {
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
	t.Properties.Set(name, schema)

	if required {
		t.Required = append(t.Required, name)
	}
}
