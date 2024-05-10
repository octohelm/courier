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
	Type  string `json:"type,omitempty"`
	Items Schema `json:"items,omitempty"`

	// validate
	MaxItems    *uint64 `json:"maxItems,omitempty"`
	MinItems    *uint64 `json:"minItems,omitempty"`
	UniqueItems *bool   `json:"uniqueItems,omitempty"`

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
