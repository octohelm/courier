package jsonschema

import "io"

func Any() *AnyType {
	return &AnyType{}
}

type AnyType struct {
	Core
	Metadata
}

func (AnyType) PrintTo(w io.Writer) {
	Print(w, func(p Printer) {
		p.Print("_")
	})
}
