package jsonschema

import "io"

func Any() *AnyType {
	a := &AnyType{}
	a.AddExtension(XGoType, "any")
	return a
}

type AnyType struct {
	Core
	Metadata
}

func (AnyType) PrintTo(w io.Writer, optionFns ...SchemaPrintOption) {
	Print(w, func(p Printer) {
		p.Print("_")
	})
}
