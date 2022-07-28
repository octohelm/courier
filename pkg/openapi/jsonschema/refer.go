package jsonschema

import (
	"bytes"
	"encoding/json"
	"strings"
)

type Refer interface {
	RefString() string
}

func RefSchemaByRefer(refer Refer) *Schema {
	return &Schema{
		Reference: Reference{
			Refer: refer,
		},
	}
}

func RefSchema(ref string) *Schema {
	return RefSchemaByRefer(ParseRef(ref))
}

type Reference struct {
	Refer
}

func (ref Reference) MarshalJSONRefFirst(values ...any) ([]byte, error) {
	if ref.Refer != nil {
		return json.Marshal(&struct {
			Ref string `json:"$ref,omitempty"`
		}{
			Ref: ref.RefString(),
		})
	}
	return FlattenMarshalJSON(values...)
}

func (ref *Reference) UnmarshalJSONRefFirst(data []byte, values ...any) error {
	r := struct {
		Ref Ref `json:"$ref,omitempty"`
	}{}

	if err := json.Unmarshal(data, &r); err != nil {
		return err
	}

	if !r.Ref.IsZero() {
		ref.Refer = &r.Ref
		return nil
	}
	// if ref exists, should ignore others
	return FlattenUnmarshalJSON(data, values...)
}

func ParseRef(ref string) *Ref {
	parts := strings.Split(ref, "#")

	r := &Ref{}

	path := ""

	if len(parts) >= 2 {
		r.Remote = parts[0]
		path = parts[1]
	} else {
		path = parts[0]
	}

	if len(path) > 0 {
		if path[0] == '/' {
			r.Paths = strings.Split(path[1:], "/")
		}
	} else {
		r.RootSelf = true
	}

	return r
}

type Ref struct {
	Remote   string
	Paths    []string
	RootSelf bool
}

func (r Ref) IsZero() bool {
	return len(r.Paths) == 0 && !r.RootSelf
}

func (r Ref) MarshalText() (text []byte, err error) {
	return []byte(r.RefString()), nil
}

func (r *Ref) UnmarshalText(text []byte) (err error) {
	ref := ParseRef(string(text))
	if !ref.IsZero() {
		*r = *ref
	}
	return nil
}

func (r Ref) RefString() string {
	b := bytes.NewBuffer(nil)

	if r.Remote != "" {
		b.WriteString(r.Remote)
	}

	b.WriteRune('#')

	for _, p := range r.Paths {
		b.WriteRune('/')
		b.WriteString(p)
	}

	return b.String()
}
