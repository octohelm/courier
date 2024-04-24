package util

import (
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/go-json-experiment/json"
	"github.com/octohelm/courier/pkg/openapi/jsonschema"
	testingx "github.com/octohelm/x/testing"
)

func TestUnmarshalTaggedUnionFromJSON(t *testing.T) {
	p := &Payload{}

	t.Run("UnmarshalJSON ok", func(t *testing.T) {
		err := p.UnmarshalJSON([]byte(`{ "kind": "String", "value": "s" }`))
		testingx.Expect(t, err, testingx.Be[error](nil))

		err = p.UnmarshalJSON([]byte(`{ "kind": "Bool", "value": true }`))
		testingx.Expect(t, err, testingx.Be[error](nil))
	})

	t.Run("UnmarshalJSON failed", func(t *testing.T) {
		err := p.UnmarshalJSON([]byte(`{ "kind": "String", "value": true }`))
		spew.Dump(err.Error())
	})
}

type Payload struct {
	Underlying any `json:"-"`
}

func (m *Payload) UnmarshalJSON(data []byte) error {
	mm := Payload{}
	if err := UnmarshalTaggedUnionFromJSON(data, &mm); err != nil {
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

var _ jsonschema.GoTaggedUnionType = &Payload{}

type TypeA struct {
	Kind  string `json:"kind"`
	Value string `json:"value"`
}

type TypeB struct {
	Kind  string `json:"kind"`
	Value bool   `json:"value"`
}
