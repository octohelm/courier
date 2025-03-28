package validator

import (
	"github.com/octohelm/courier/pkg/validator/internal"
	"github.com/octohelm/courier/pkg/validator/internal/rules"

	_ "github.com/octohelm/courier/pkg/validator/strfmt"
	_ "github.com/octohelm/courier/pkg/validator/validators"
)

type (
	Creator = internal.ValidatorProvider
	Option  = internal.ValidatorOption
)

type WithDefaultValue = internal.WithDefaultValue

type WithStructTagValidate = internal.WithStructTagValidate

type Validator = internal.Validator

func Register(creator Creator) {
	internal.Register(creator)
}

func New(option Option) (Validator, error) {
	return internal.New(option)
}

func NewFormatValidatorProvider(format string, createValidator func(format string) Validator) internal.ValidatorProvider {
	return internal.CreateValidatorProvider([]string{format}, func(rule *rules.Rule) (Validator, error) {
		return createValidator(format), nil
	})
}
