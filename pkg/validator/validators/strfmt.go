package validators

import (
	"regexp"

	"github.com/octohelm/courier/pkg/validator/internal/rules"

	"github.com/go-json-experiment/json/jsontext"
	validatorerrors "github.com/octohelm/courier/pkg/validator/errors"
	"github.com/octohelm/courier/pkg/validator/internal"
)

func NewRegexpStrfmtValidatorProvider(regexpStr string, name string, aliases ...string) internal.ValidatorProvider {
	re := regexp.MustCompile(regexpStr)

	validate := func(s string) error {
		if !re.MatchString(s) {
			return &validatorerrors.ErrNotMatch{
				Topic:   name,
				Current: s,
				Pattern: re.String(),
			}
		}
		return nil
	}

	return NewStrfmtValidatorProvider(validate, name, aliases...)
}

func NewStrfmtValidatorProvider(validate func(unquoteStr string) error, name string, aliases ...string) internal.ValidatorProvider {
	return &strfmtValidatorProvider{
		names:    append([]string{name}, aliases...),
		validate: validate,
	}
}

type strfmtValidatorProvider struct {
	names    []string
	validate func(unquoteStr string) error
}

func (s *strfmtValidatorProvider) Names() []string {
	return s.names
}

func (s *strfmtValidatorProvider) Validator(rule *rules.Rule) (internal.Validator, error) {
	return &StrfmtValidator{
		name:     rule.Name,
		validate: s.validate,
	}, nil
}

type StrfmtValidator struct {
	name     string
	validate func(unquoteStr string) error
}

func (validator *StrfmtValidator) Format() string {
	return validator.name
}

func (validator *StrfmtValidator) String() string {
	return "@" + validator.name
}

func (validator *StrfmtValidator) Validate(value jsontext.Value) error {
	if value.Kind() != '"' {
		return &validatorerrors.ErrInvalidType{
			Type:  "string",
			Value: string(value),
		}
	}

	unquote, err := jsontext.AppendUnquote(nil, value)
	if err != nil {
		return err
	}

	val := string(unquote)

	if err := validator.validate(val); err != nil {
		return err
	}

	return nil
}
