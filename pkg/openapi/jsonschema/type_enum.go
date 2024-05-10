package jsonschema

import (
	"io"
	"reflect"
)

type GoEnumValues interface {
	EnumValues() []any
}

type EnumType struct {
	Enum []any `json:"enum"`

	Core
	Metadata
}

func (t EnumType) PrintTo(w io.Writer, optionFns ...SchemaPrintOption) {
	Print(w, func(p Printer) {
		for i, v := range t.Enum {
			if i > 0 {
				p.Printf(" | ")
			}
			
			switch reflect.TypeOf(v).Kind() {
			case reflect.String:
				p.Printf("%q", v)
			default:
				p.Printf("%v", v)
			}
		}
	})
}
