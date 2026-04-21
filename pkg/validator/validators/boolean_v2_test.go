package validators

import (
	"regexp"
	"testing"

	"github.com/go-json-experiment/json/jsontext"
	"github.com/octohelm/courier/pkg/validator/internal/rules"
	. "github.com/octohelm/x/testing/v2"
)

func TestBooleanValidator(t0 *testing.T) {
	provider := &booleanValidatorProvider{}

	Then(t0, "boolean validator provider 与 validator 行为正确",
		Expect(provider.Names(), Equal([]string{"bool", "boolean"})),
		ExpectMust(func() error {
			v, err := provider.Validator(rules.MustParseRuleString("@bool"))
			if err != nil {
				return err
			}
			if v.String() != "@bool" {
				return errBool("unexpected validator string")
			}
			if err := v.Validate(jsontext.Value(`true`)); err != nil {
				return err
			}
			if err := v.Validate(jsontext.Value(`false`)); err != nil {
				return err
			}
			return nil
		}),
		ExpectDo(func() error {
			v, _ := provider.Validator(rules.MustParseRuleString("@bool"))
			return v.Validate(jsontext.Value(`1`))
		}, ErrorMatch(regexp.MustCompile("invalid boolean"))),
	)
}

func TestValidatorsRuntimeDoc(t0 *testing.T) {
	Then(t0, "生成的 runtime doc 可以访问",
		ExpectMust(func() error {
			if doc, ok := (&FloatValidator{}).RuntimeDoc(); !ok || len(doc) == 0 {
				return errBool("missing float runtime doc")
			}
			for _, name := range []string{"BitSize", "MaxDigits", "DecimalDigits"} {
				if _, ok := (&FloatValidator{}).RuntimeDoc(name); !ok {
					return errBool("missing float runtime field doc")
				}
			}
			if _, ok := (&FloatValidator{}).RuntimeDoc("Minimum"); !ok {
				return errBool("missing nested float number runtime doc")
			}
			return nil
		}),
		ExpectMust(func() error {
			if doc, ok := (&IntegerValidator[int64]{}).RuntimeDoc(); !ok || len(doc) == 0 {
				return errBool("missing integer runtime doc")
			}
			for _, name := range []string{"Unsigned", "BitSize", "Maximum"} {
				if _, ok := (&IntegerValidator[int64]{}).RuntimeDoc(name); !ok {
					return errBool("missing integer runtime field doc")
				}
			}
			return nil
		}),
		ExpectMust(func() error {
			if doc, ok := (&MapValidator{}).RuntimeDoc(); !ok || len(doc) == 0 {
				return errBool("missing map runtime doc")
			}
			for _, name := range []string{"MinProperties", "MaxProperties"} {
				if _, ok := (&MapValidator{}).RuntimeDoc(name); !ok {
					return errBool("missing map runtime field doc")
				}
			}
			return nil
		}),
		ExpectMust(func() error {
			if doc, ok := (&Number[int64]{}).RuntimeDoc(); !ok || len(doc) != 0 {
				return errBool("unexpected number runtime doc")
			}
			for _, name := range []string{"Minimum", "Maximum", "MultipleOf", "ExclusiveMaximum", "ExclusiveMinimum", "Enums"} {
				if _, ok := (&Number[int64]{}).RuntimeDoc(name); !ok {
					return errBool("missing number runtime field doc")
				}
			}
			return nil
		}),
		ExpectMust(func() error {
			if doc, ok := (&SliceValidator{}).RuntimeDoc(); !ok || len(doc) == 0 {
				return errBool("missing slice runtime doc")
			}
			for _, name := range []string{"MinItems", "MaxItems"} {
				if _, ok := (&SliceValidator{}).RuntimeDoc(name); !ok {
					return errBool("missing slice runtime field doc")
				}
			}
			if doc, ok := new(StrLenMode).RuntimeDoc(); !ok || len(doc) != 0 {
				return errBool("unexpected strlen runtime doc")
			}
			return nil
		}),
		ExpectMust(func() error {
			if doc, ok := (&StringValidator{}).RuntimeDoc(); !ok || len(doc) == 0 {
				return errBool("missing string runtime doc")
			}
			for _, name := range []string{"Pattern", "LenMode", "MinLength", "MaxLength", "Enums"} {
				if _, ok := (&StringValidator{}).RuntimeDoc(name); !ok {
					return errBool("missing string runtime field doc")
				}
			}
			if doc, ok := runtimeDoc(&StringValidator{}, "prefix:", "Pattern"); !ok || len(doc) != 0 {
				return errBool("unexpected prefixed runtime doc")
			}
			if _, ok := runtimeDoc(struct{}{}, "", "Pattern"); ok {
				return errBool("unexpected runtimeDoc hit")
			}
			return nil
		}),
	)
}

func errBool(msg string) error {
	return &boolErr{msg: msg}
}

type boolErr struct {
	msg string
}

func (e *boolErr) Error() string {
	return e.msg
}
