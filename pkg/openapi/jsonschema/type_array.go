package jsonschema

import (
	"io"
)

func ArrayOf(items Schema) *ArrayType {
	return &ArrayType{
		Type:  "array",
		Items: items,
	}
}

type ArrayType struct {
	Type  string `json:"type,omitzero"`
	Items Schema `json:"items,omitzero"`

	// validate
	MaxItems    *uint64 `json:"maxItems,omitzero"`
	MinItems    *uint64 `json:"minItems,omitzero"`
	UniqueItems *bool   `json:"uniqueItems,omitzero"`

	Core
	Metadata
}

func (s ArrayType) PrintTo(w io.Writer, optionFns ...SchemaPrintOption) {
	Print(w, func(p Printer) {
		p.Print("[...")
		p.PrintFrom(s.Items, optionFns...)
		p.Print("]")
	})
}
