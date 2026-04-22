package validator

import (
	"bytes"
	"fmt"
	"reflect"
	"testing"

	"github.com/go-json-experiment/json/jsontext"
	. "github.com/octohelm/x/testing/v2"

	"github.com/octohelm/courier/pkg/validator/internal/rules"
)

type formatValidator struct{}

func (formatValidator) Validate(jsontext.Value) error { return nil }
func (formatValidator) String() string                { return "@custom-format" }

func TestMarshalAndUnmarshalHelpers(t0 *testing.T) {
	type payload struct {
		Name  string `json:"name"`
		Empty string `json:"empty,omitempty"`
	}

	Then(t0, "公开 marshal 与 unmarshal 入口可正常工作",
		ExpectMust(func() error {
			value := &payload{}
			if err := UnmarshalRead(bytes.NewBufferString(`{"name":"demo"}`), value); err != nil {
				return err
			}
			if value.Name != "demo" {
				return errf("unexpected value %#v", value)
			}
			return nil
		}),
		ExpectMust(func() error {
			value := &payload{}
			if err := UnmarshalDecode(jsontext.NewDecoder(bytes.NewBufferString(`{"name":"decode"}`)), value); err != nil {
				return err
			}
			if value.Name != "decode" {
				return errf("unexpected value %#v", value)
			}
			return nil
		}),
		ExpectMust(func() error {
			buf := bytes.NewBuffer(nil)
			if err := MarshalWrite(buf, payload{Name: "demo"}); err != nil {
				return err
			}
			if buf.String() != `{"name":"demo"}` {
				return errf("unexpected json %s", buf.String())
			}
			return nil
		}),
		ExpectMust(func() error {
			buf := bytes.NewBuffer(nil)
			if err := MarshalEncode(jsontext.NewEncoder(buf), payload{Name: "encode"}); err != nil {
				return err
			}
			if bytes.TrimSpace(buf.Bytes()) == nil {
				return errf("empty buffer")
			}
			return nil
		}),
		ExpectMust(func() error {
			data, err := Marshal(payload{Name: "marshal"})
			if err != nil {
				return err
			}
			if string(data) != `{"name":"marshal"}` {
				return errf("unexpected json %s", string(data))
			}
			return nil
		}),
	)
}

func TestValidatorProviderHelpers(t0 *testing.T) {
	provider := NewFormatValidatorProvider("custom-format", func(string) Validator {
		return formatValidator{}
	})
	Register(provider)

	Then(t0, "validator provider 包装器可正确注册并创建校验器",
		ExpectMust(func() error {
			if !reflect.DeepEqual(provider.Names(), []string{"custom-format"}) {
				return errf("unexpected names %v", provider.Names())
			}
			v, err := provider.Validator(rules.MustParseRuleString("@custom-format"))
			if err != nil {
				return err
			}
			if v.String() != "@custom-format" {
				return errf("unexpected validator %s", v.String())
			}
			return nil
		}),
		ExpectMust(func() error {
			opt := Option{
				Type:     reflect.TypeFor[string](),
				Rule:     "@custom-format",
				Optional: true,
				String:   true,
			}
			if err := opt.SetDefaultValue("demo"); err != nil {
				return err
			}
			v, err := New(opt)
			if err != nil {
				return err
			}
			if v.String() != "@custom-format?" {
				return errf("unexpected validator string %s", v.String())
			}
			if withDefault, ok := v.(WithDefaultValue); !ok || string(withDefault.DefaultValue()) != `"demo"` {
				return errf("unexpected default value")
			}
			return nil
		}),
	)
}

func errf(format string, args ...any) error {
	return &validatorHelperError{msg: fmt.Sprintf(format, args...)}
}

type validatorHelperError struct {
	msg string
}

func (e *validatorHelperError) Error() string {
	return e.msg
}
