package jsonschema

import (
	"io"

	validatortaggedunion "github.com/octohelm/courier/pkg/validator/taggedunion"
)

type GoTaggedUnionType = validatortaggedunion.Type

type GoUnionType interface {
	OneOf() []any
}

func OneOf(schemas ...Schema) *UnionType {
	return &UnionType{
		OneOf: schemas,
	}
}

type UnionType struct {
	OneOf         []Schema       `json:"oneOf"`
	Discriminator *Discriminator `json:"discriminator,omitzero"`

	Core
	Metadata
}

type Discriminator struct {
	PropertyName string            `json:"propertyName"`
	Mapping      map[string]Schema `json:"mapping,omitzero"`
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
