package jsonschema

import (
	"io"
)

func AllOf(schemas ...Schema) *IntersectionType {
	return &IntersectionType{
		AllOf: schemas,
	}
}

type IntersectionType struct {
	AllOf []Schema `json:"allOf"`

	Core
	Metadata
}

func (t IntersectionType) PrintTo(w io.Writer, optionFns ...SchemaPrintOption) {
	Print(w, func(p Printer) {
		for i, s := range t.AllOf {
			// skip any
			if _, ok := s.(*AnyType); ok {
				continue
			}

			if i > 0 {
				p.Print(" & ")
			}
			p.PrintFrom(s, optionFns...)
		}
	})
}
