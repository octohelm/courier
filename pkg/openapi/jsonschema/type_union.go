package jsonschema

import (
	"io"
)

type GoUnionType interface {
	OneOf() []any
}

type GoTaggedUnionType interface {
	Discriminator() string
	Mapping() map[string]any
	SetUnderlying(u any)
}

func OneOf(schemas ...Schema) *UnionType {
	return &UnionType{
		OneOf: schemas,
	}
}

type UnionType struct {
	OneOf         []Schema       `json:"oneOf"`
	Discriminator *Discriminator `json:"discriminator,omitempty"`

	Core
	Metadata
}

type Discriminator struct {
	PropertyName string            `json:"propertyName"`
	Mapping      map[string]Schema `json:"mapping,omitempty"`
}

func (t UnionType) PrintTo(w io.Writer, optionFns ...SchemaPrintOption) {
	Print(w, func(p Printer) {
		for i, s := range t.OneOf {
			if i > 0 {
				p.Print(" | ")
			}
			p.PrintFrom(s)
		}
	})
}
