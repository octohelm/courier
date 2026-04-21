package validators

import (
	"errors"
	"regexp"
	"testing"

	"github.com/go-json-experiment/json/jsontext"
	"github.com/octohelm/courier/pkg/validator/internal/rules"
	. "github.com/octohelm/x/testing/v2"
)

func TestStrfmtValidatorProviderAndValidator(t0 *testing.T) {
	provider := NewStrfmtValidatorProvider(func(s string) error {
		if s != "ok" {
			return errors.New("not ok")
		}
		return nil
	}, "demo", "alias")

	Then(t0, "strfmt validator provider 与 validator 行为正确",
		Expect(provider.Names(), Equal([]string{"demo", "alias"})),
		ExpectMust(func() error {
			v, err := provider.Validator(rules.MustParseRuleString("@demo"))
			if err != nil {
				return err
			}
			if x, ok := v.(*StrfmtValidator); !ok || x.Format() != "demo" || x.String() != "@demo" {
				return errValidator("unexpected validator")
			}
			if err := v.Validate(jsontext.Value(`"ok"`)); err != nil {
				return err
			}
			return nil
		}),
		ExpectDo(func() error {
			v, _ := provider.Validator(rules.MustParseRuleString("@demo"))
			return v.Validate(jsontext.Value(`1`))
		}, ErrorMatch(mustValidatorRE("invalid string"))),
		ExpectDo(func() error {
			v, _ := provider.Validator(rules.MustParseRuleString("@demo"))
			return v.Validate(jsontext.Value(`"bad"`))
		}, ErrorMatch(mustValidatorRE("not ok"))),
	)
}

func TestRegexpStrfmtValidatorProvider(t0 *testing.T) {
	provider := NewRegexpStrfmtValidatorProvider(`^[0-9]+$`, "numeric-demo")

	Then(t0, "regexp strfmt validator provider 可匹配并返回模式错误",
		ExpectMust(func() error {
			v, err := provider.Validator(rules.MustParseRuleString("@numeric-demo"))
			if err != nil {
				return err
			}
			return v.Validate(jsontext.Value(`"123"`))
		}),
		ExpectDo(func() error {
			v, _ := provider.Validator(rules.MustParseRuleString("@numeric-demo"))
			return v.Validate(jsontext.Value(`"abc"`))
		}, ErrorMatch(mustValidatorRE(`numeric-demo should match`))),
	)
}

func mustValidatorRE(s string) *regexp.Regexp {
	return regexp.MustCompile(s)
}

func errValidator(msg string) error {
	return &validatorErr{msg: msg}
}

type validatorErr struct {
	msg string
}

func (e *validatorErr) Error() string {
	return e.msg
}
