package jsonschema

import "io"

type NullType struct {
	Type string `json:"type,omitzero"`

	Core
	Metadata
}

func (NullType) PrintTo(w io.Writer, optionFns ...SchemaPrintOption) {
	Print(w, func(p Printer) {
		p.Print("null")
	})
}
