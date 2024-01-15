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
	Type   string `json:"type,omitempty"`
	Format string `json:"format,omitempty"`

	// validate
	MaxLength *uint64 `json:"maxLength,omitempty"`
	MinLength *uint64 `json:"minLength,omitempty"`
	Pattern   string  `json:"pattern,omitempty"`

	Core
	Metadata
}

func (t StringType) PrintTo(w io.Writer) {
	Print(w, func(p Printer) {
		if t.Format == "bytes" || t.Format == "binary" {
			p.Print("[]byte")
			return
		}
		p.Print("string")
	})
}
