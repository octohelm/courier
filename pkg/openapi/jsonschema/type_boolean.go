package jsonschema

import "io"

func Boolean() *BooleanType {
	return &BooleanType{
		Type: "boolean",
	}
}

type BooleanType struct {
	Type string `json:"type,omitempty"`

	Core
	Metadata
}

func (BooleanType) PrintTo(w io.Writer) {
	Print(w, func(p Printer) {
		p.Print("bool")
	})
}
