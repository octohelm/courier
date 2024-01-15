package jsonschema

import (
	"io"
)

type GoEnumValues interface {
	EnumValues() []any
}

type EnumType struct {
	Enum []any `json:"enum"`

	Core
	Metadata
}

func (t EnumType) PrintTo(w io.Writer) {
	Print(w, func(p Printer) {
		for i, v := range t.Enum {
			if i > 0 {
				p.Printf(" | ")
			}
			p.Printf("%v", v)
		}
	})
}
