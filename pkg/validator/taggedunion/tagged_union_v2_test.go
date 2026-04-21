package taggedunion

import (
	"bytes"
	"regexp"
	"testing"

	"github.com/go-json-experiment/json/jsontext"
	. "github.com/octohelm/x/testing/v2"
)

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
