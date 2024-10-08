package validator

import (
	"github.com/octohelm/courier/pkg/validator/internal"

	_ "github.com/octohelm/courier/pkg/validator/strfmt"
	_ "github.com/octohelm/courier/pkg/validator/validators"
)

type Creator = internal.ValidatorProvider
type Option = internal.ValidatorOption

type Validator = internal.Validator

func Register(creator Creator) {
	internal.Register(creator)
}

func New(option Option) (Validator, error) {
	return internal.New(option)
}
