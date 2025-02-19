package validators

import (
	"github.com/go-json-experiment/json/jsontext"
	validatorerrors "github.com/octohelm/courier/pkg/validator/errors"
	"github.com/octohelm/courier/pkg/validator/internal"
	"github.com/octohelm/courier/pkg/validator/internal/rules"
)

func init() {
	internal.Register(&booleanValidatorProvider{})
}

type booleanValidatorProvider struct{}

func (booleanValidatorProvider) Names() []string {
	return []string{
		"bool", "boolean",
	}
}

func (c *booleanValidatorProvider) Validator(rule *rules.Rule) (internal.Validator, error) {
	return &booleanValidator{}, nil
}

type booleanValidator struct{}

func (validator *booleanValidator) String() string {
	return "@bool"
}

func (validator *booleanValidator) Validate(value jsontext.Value) error {
	if !(value.Kind() == 'f' || value.Kind() == 't') {
		return &validatorerrors.ErrInvalidType{
			Type:  "boolean",
			Value: string(value),
		}
	}
	return nil
}
