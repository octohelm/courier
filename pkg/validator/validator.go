package validator

import (
	"github.com/octohelm/courier/pkg/validator/internal"

	_ "github.com/octohelm/courier/pkg/validator/strfmt"
	_ "github.com/octohelm/courier/pkg/validator/validators"
)

type (
	Creator = internal.ValidatorProvider
	Option  = internal.ValidatorOption
)

type WithDefaultValue = internal.WithDefaultValue

type Validator = internal.Validator

func Register(creator Creator) {
	internal.Register(creator)
}

func New(option Option) (Validator, error) {
	return internal.New(option)
}
