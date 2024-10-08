package jsonschema

import "io"

func String() *StringType {
	return &StringType{
		Type: "string",
	}
}

func Bytes() *StringType {
	return &StringType{
		Type:   "string",
		Format: "bytes",
	}
}

func Binary() *StringType {
	return &StringType{
		Type:   "string",
		Format: "binary",
	}
}

type StringType struct {
	Type   string `json:"type,omitzero"`
	Format string `json:"format,omitzero"`

	// validate
	MaxLength *uint64 `json:"maxLength,omitzero"`
	MinLength *uint64 `json:"minLength,omitzero"`
	Pattern   string  `json:"pattern,omitzero"`

	Core
	Metadata
}

func (t StringType) PrintTo(w io.Writer, optionFns ...SchemaPrintOption) {
	Print(w, func(p Printer) {
		if t.Format == "bytes" || t.Format == "binary" {
			p.Print("[]byte")
			return
		}
		p.Print("string")
	})
}
