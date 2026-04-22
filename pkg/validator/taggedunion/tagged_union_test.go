package taggedunion

import (
	"bytes"
	"fmt"
	"regexp"
	"testing"

	"github.com/go-json-experiment/json"
	"github.com/go-json-experiment/json/jsontext"
	. "github.com/octohelm/x/testing/v2"

	"github.com/octohelm/courier/pkg/validator"
)

func TestUnmarshalTaggedUnionFromJSON(t *testing.T) {
	t.Run("string discriminator", func(t *testing.T) {
		Then(t, "String kind 会解码到对应实现", ExpectMust(func() error {
			p := &Payload{}
			if err := validator.Unmarshal([]byte(`{ "kind": "String", "value": "s" }`), p); err != nil {
				return err
			}
			got, ok := p.Underlying.(*StringKinded)
			if !ok || *got != (StringKinded{Kind: "String", Value: "s"}) {
				return fmt.Errorf("unexpected string discriminator result: %#v", p.Underlying)
			}
			return nil
		}))
	})

	t.Run("bool discriminator", func(t *testing.T) {
		Then(t, "Bool kind 会解码到对应实现", ExpectMust(func() error {
			p := &Payload{}
			if err := validator.Unmarshal([]byte(`{ "kind": "Bool", "value": true }`), p); err != nil {
				return err
			}
			got, ok := p.Underlying.(*BoolKinded)
			if !ok || *got != (BoolKinded{Kind: "Bool", Value: true}) {
				return fmt.Errorf("unexpected bool discriminator result: %#v", p.Underlying)
			}
			return nil
		}))
	})

	t.Run("invalid nested value", func(t *testing.T) {
		Then(t, "嵌套路径中的类型不匹配会保留完整 pointer",
			ExpectDo(func() error {
				p := &struct {
					Path struct {
						To Payload `json:"to"`
					} `json:"path"`
				}{}

				return validator.Unmarshal([]byte(`{ "path": { "to": { "kind": "String", "value": true } } } }`), p)
			}, ErrorMatch(mustTaggedUnionRE(`^invalid string: true at /path/to/value$`))),
		)
	})
}

func TestTaggedUnionBranches(t0 *testing.T) {
	Then(t0, "tagged union 补齐空值与错误分支",
		ExpectMust(func() error {
			p := &Payload{}
			if err := Unmarshal([]byte(`null`), p); err != nil {
				return err
			}
			if p.Underlying != nil {
				return errTagged("unexpected underlying value")
			}
			return nil
		}),
		ExpectMust(func() error {
			p := &Payload{}
			if err := Unmarshal([]byte(`{"value":"s"}`), p); err != nil {
				return err
			}
			if p.Underlying != nil {
				return errTagged("unexpected underlying when discriminator missing")
			}
			return nil
		}),
		ExpectDo(func() error {
			p := &Payload{}
			return Unmarshal([]byte(`{"kind":"Unknown","value":"x"}`), p)
		}, ErrorMatch(mustTaggedRE(`unsupported kind=Unknown at /kind`))),
		ExpectDo(func() error {
			p := &Payload{}
			return UnmarshalDecode(jsontext.NewDecoder(bytes.NewBufferString(`{`)), p)
		}, ErrorMatch(mustTaggedRE(`EOF|unexpected`))),
		ExpectDo(func() error {
			p := &Payload{}
			return UnmarshalDecode(jsontext.NewDecoder(bytes.NewBufferString(``)), p)
		}, ErrorMatch(mustTaggedRE(`EOF|unexpected`))),
		ExpectDo(func() error {
			p := &Payload{}
			return UnmarshalDecode(jsontext.NewDecoder(bytes.NewBufferString(`1`)), p)
		}, ErrorMatch(mustTaggedRE(`invalid object`))),
	)
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

func mustTaggedUnionRE(pattern string) *regexp.Regexp {
	return regexp.MustCompile(pattern)
}

func mustTaggedRE(s string) *regexp.Regexp {
	return regexp.MustCompile(s)
}

func errTagged(msg string) error {
	return &taggedErr{msg: msg}
}

type taggedErr struct {
	msg string
}

func (e *taggedErr) Error() string {
	return e.msg
}
