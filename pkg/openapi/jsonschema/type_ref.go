package jsonschema

import "io"

type RefType struct {
	Ref        *URIReferenceString `json:"$ref,omitempty"`
	DynamicRef *URIReferenceString `json:"$dynamicRef,omitempty"`

	Core
	Metadata
}

func (t RefType) RefName() string {
	if t.Ref != nil {
		return t.Ref.RefName()
	}

	if t.DynamicRef != nil {
		return t.DynamicRef.RefName()
	}

	return "invalid"
}

func (t RefType) PrintTo(w io.Writer) {
	Print(w, func(p Printer) {
		if t.Ref != nil {
			p.Printf("#%s", t.Ref.RefName())
			return
		}

		if t.DynamicRef != nil {
			p.Printf("#%s", t.DynamicRef.RefName())
			return
		}

		p.Printf("_|_")

		return
	})
}
