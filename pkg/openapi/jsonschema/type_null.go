package jsonschema

import "io"

type NullType struct {
	Type string `json:"type,omitempty"`

	Core
	Metadata
}

func (NullType) PrintTo(w io.Writer, optionFns ...SchemaPrintOption) {
	Print(w, func(p Printer) {
		p.Print("null")
	})
}
