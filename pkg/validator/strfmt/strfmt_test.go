package strfmt

import (
	"testing"

	"github.com/go-json-experiment/json/jsontext"
	"github.com/octohelm/courier/pkg/validator/internal/rules"
	. "github.com/octohelm/x/testing/v2"
)

func TestBuiltinStrfmtProviders(t0 *testing.T) {
	Then(t0, "内置字符串格式校验 provider 可工作",
		ExpectMust(func() error {
			if len(ASCIIValidatorProvider.Names()) == 0 || ASCIIValidatorProvider.Names()[0] != "ascii" {
				return errStrfmt("unexpected ascii provider names")
			}
			v, err := ASCIIValidatorProvider.Validator(rules.MustParseRuleString("@ascii"))
			if err != nil {
				return err
			}
			if err := v.Validate(jsontext.Value(`"hello"`)); err != nil {
				return err
			}
			return nil
		}),
		ExpectMust(func() error {
			if len(AlphaNumericValidatorProvider.Names()) != 2 {
				return errStrfmt("unexpected alpha numeric aliases")
			}
			v, err := EmailValidatorProvider.Validator(rules.MustParseRuleString("@email"))
			if err != nil {
				return err
			}
			if err := v.Validate(jsontext.Value(`"demo@example.com"`)); err != nil {
				return err
			}
			return nil
		}),
	)
}

func errStrfmt(msg string) error {
	return &strfmtErr{msg: msg}
}

type strfmtErr struct {
	msg string
}

func (e *strfmtErr) Error() string {
	return e.msg
}
