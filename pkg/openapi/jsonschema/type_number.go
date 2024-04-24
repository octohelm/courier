package jsonschema

import (
	"io"

	"github.com/octohelm/courier/pkg/ptr"
)

func Integer() Schema {
	t := &NumberType{
		Type:    "integer",
		Minimum: ptr.Ptr(float64(-1 << (32 - 1))),
		Maximum: ptr.Ptr(float64(1<<(32-1) - 1)),
	}

	t.AddExtension("x-format", "int32")
	return t
}

func Long() Schema {
	t := &NumberType{
		Type: "integer",
	}
	t.AddExtension("x-format", "int64")
	return t
}

type NumberType struct {
	Type string `json:"type,omitempty"`

	// validate
	MultipleOf       *float64 `json:"multipleOf,omitempty"`
	Maximum          *float64 `json:"maximum,omitempty"`
	ExclusiveMaximum *float64 `json:"exclusiveMaximum,omitempty"`
	Minimum          *float64 `json:"minimum,omitempty"`
	ExclusiveMinimum *float64 `json:"exclusiveMinimum,omitempty"`

	Core
	Metadata
}

func (t NumberType) PrintTo(w io.Writer) {
	Print(w, func(p Printer) {
		if t.Type == "integer" {
			p.Print("int")
			return
		}
		p.Print("number")
	})
}
