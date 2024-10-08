package taggedunion

import (
	"bytes"
	"testing"

	"github.com/go-json-experiment/json/jsontext"
	"github.com/octohelm/courier/pkg/validator"

	"github.com/davecgh/go-spew/spew"
	"github.com/go-json-experiment/json"
	testingx "github.com/octohelm/x/testing"
)

func TestUnmarshalTaggedUnionFromJSON(t *testing.T) {
	t.Run("UnmarshalJSON ok", func(t *testing.T) {
		p := &Payload{}
		err := validator.Unmarshal([]byte(`{ "kind": "String", "value": "s" }`), p)
		testingx.Expect(t, err, testingx.Be[error](nil))

		err = validator.Unmarshal([]byte(`{ "kind": "Bool", "value": true }`), p)
		testingx.Expect(t, err, testingx.Be[error](nil))
	})

	t.Run("UnmarshalJSON failed", func(t *testing.T) {
		p := &Payload{}
		err := validator.Unmarshal([]byte(`{ "kind": "String", "value": true }`), p)
		spew.Dump(err)
	})

	t.Run("UnmarshalJSON failed in depth", func(t *testing.T) {
		p := &struct {
			Path struct {
				To Payload `json:"to"`
			} `json:"path"`
		}{}

		err := validator.Unmarshal([]byte(`{ "path": { "to": { "kind": "String", "value": true } }}`), p)

		testingx.Expect(t, err.Error(), testingx.Be("invalid string: true at /path/to/value"))
	})
}

type Payload struct {
	Underlying any `json:"-"`
}

func (m *Payload) UnmarshalJSON(data []byte) error {
	mm := Payload{}
	if err := UnmarshalDecode(jsontext.NewDecoder(bytes.NewBuffer(data)), &mm); err != nil {
		return err
	}
	*m = mm
	return nil
}

func (m Payload) MarshalJSON() ([]byte, error) {
	return json.Marshal(m.Underlying)
}

func (p Payload) Discriminator() string {
	return "kind"
}

func (p Payload) Mapping() map[string]any {
	return map[string]any{
		"String": &TypeA{},
		"Bool":   &TypeB{},
	}
}

func (p *Payload) SetUnderlying(u any) {
	p.Underlying = u
}

var _ Type = &Payload{}

type TypeA struct {
	Kind  string `json:"kind"`
	Value string `json:"value"`
}

type TypeB struct {
	Kind  string `json:"kind"`
	Value bool   `json:"value"`
}
