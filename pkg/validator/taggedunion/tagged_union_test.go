package taggedunion

import (
	"bytes"
	"testing"

	"github.com/go-json-experiment/json"
	"github.com/go-json-experiment/json/jsontext"
	"github.com/octohelm/courier/pkg/validator"
	"github.com/octohelm/x/testing/bdd"
)

func TestUnmarshalTaggedUnionFromJSON(t *testing.T) {
	b := bdd.FromT(t)

	b.Given("object", func(b bdd.T) {
		b.When("unmarshal for kind String", func(b bdd.T) {
			p := &Payload{}

			b.Then("success",
				bdd.NoError(validator.Unmarshal([]byte(`{ "kind": "String", "value": "s" }`), p)),
			)

			b.Then("value",
				bdd.Equal(StringKinded{Kind: "String", Value: "s"}, *p.Underlying.(*StringKinded)),
			)
		})

		b.When("unmarshal for kind Bool", func(b bdd.T) {
			p := &Payload{}

			b.Then("success",
				bdd.NoError(validator.Unmarshal([]byte(`{ "kind": "Bool", "value": true }`), p)),
			)

			b.Then("value",
				bdd.Equal(BoolKinded{Kind: "Bool", Value: true}, *p.Underlying.(*BoolKinded)),
			)

			b.When("with invalid value", func(b bdd.T) {
				p := &struct {
					Path struct {
						To Payload `json:"to"`
					} `json:"path"`
				}{}

				err := validator.Unmarshal([]byte(`{ "path": { "to": { "kind": "String", "value": true } } } }`), p)
				b.Then("failed",
					bdd.Equal("invalid string: true at /path/to/value", err.Error()),
				)
			})
		})
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
		"String": &StringKinded{},
		"Bool":   &BoolKinded{},
	}
}

func (p *Payload) SetUnderlying(u any) {
	p.Underlying = u
}

var _ Type = &Payload{}

type StringKinded struct {
	Kind  string `json:"kind"`
	Value string `json:"value"`
}

type BoolKinded struct {
	Kind  string `json:"kind"`
	Value bool   `json:"value"`
}
