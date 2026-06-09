package validators

import (
	"regexp"
	"testing"

	"github.com/go-json-experiment/json/jsontext"
	. "github.com/octohelm/x/testing/v2"

	"github.com/octohelm/courier/pkg/validator/internal/rules"
)

func TestBooleanValidator(t0 *testing.T) {
	provider := &booleanValidatorProvider{}

	Then(
		t0, "boolean validator provider 与 validator 行为正确",
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

func errBool(msg string) error {
	return &boolErr{msg: msg}
}

type boolErr struct {
	msg string
}

func (e *boolErr) Error() string {
	return e.msg
}
