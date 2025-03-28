package validators

import (
	"regexp"

	validatorerrors "github.com/octohelm/courier/pkg/validator/errors"
	"github.com/octohelm/courier/pkg/validator/internal"
)

func NewRegexpStrfmtValidatorProvider(regexpStr string, name string, aliases ...string) internal.ValidatorProvider {
	re := regexp.MustCompile(regexpStr)

	validate := func(s string) error {
		if !re.MatchString(s) {
			return &validatorerrors.ErrPatternNotMatch{
				Subject: name,
				Target:  s,
				Pattern: re.String(),
			}
		}
		return nil
	}

	return NewStrfmtValidatorProvider(validate, name, aliases...)
}
