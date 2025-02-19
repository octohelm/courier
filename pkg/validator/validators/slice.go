package validators

import (
	"fmt"
	"reflect"

	"github.com/go-json-experiment/json/jsontext"
	validatorerrors "github.com/octohelm/courier/pkg/validator/errors"
	"github.com/octohelm/courier/pkg/validator/internal"
	"github.com/octohelm/courier/pkg/validator/internal/rules"
)

func init() {
	internal.Register(&sliceValidatorProvider{})
}

type sliceValidatorProvider struct{}

func (sliceValidatorProvider) Names() []string {
	return []string{"slice", "array"}
}

func (c *sliceValidatorProvider) Validator(rule *rules.Rule) (internal.Validator, error) {
	validator := &SliceValidator{}

	if rule.ExclusiveLeft || rule.ExclusiveRight {
		return nil, rules.NewSyntaxError("range mark of %s should not be `(` or `)`", c.Names()[0])
	}

	if rule.Params != nil {
		if len(rule.Params) != 1 {
			return nil, fmt.Errorf("slice should only 1 parameter, but got %d", len(rule.Params))
		}
		r, ok := rule.Params[0].(*rules.Rule)
		if !ok {
			return nil, fmt.Errorf("slice parameter should be a valid rule")
		}

		validator.elem.Rule = string(r.RAW)
	}

	if rule.Range != nil {
		if rule.Name == "array" && len(rule.Range) > 1 {
			return nil, rules.NewSyntaxError("array should declare length only")
		}

		minn, maxn, err := convertRangeValues[uint64](rule)
		if err != nil {
			return nil, err
		}
		if minn != nil {
			validator.MinItems = *minn
		}

		validator.MaxItems = maxn
	}

	return validator, nil
}

/*
Validator for slice

Rules:

	@slice<ELEM_RULE>[minLen,maxLen]
	@slice<ELEM_RULE>[length]

	@slice<@string{A,B,C}>[,100]

Aliases

	@array = @slice // and range must to be use length
*/
type SliceValidator struct {
	elem internal.ValidatorOption

	MinItems uint64
	MaxItems *uint64
}

func (validator *SliceValidator) Elem() internal.ValidatorOption {
	return validator.elem
}

func (validator *SliceValidator) Validate(value jsontext.Value) error {
	return nil
}

func (validator *SliceValidator) String() string {
	rule := rules.NewRule("slice")

	if validator.elem.Rule != "" {
		rule.Params = append(rule.Params, rules.NewRuleLit([]byte(validator.elem.Rule)))
	}

	rule.Range = make([]*rules.RuleLit, 2)

	rule.Range[0] = rules.NewRuleLit(
		[]byte(fmt.Sprintf("%d", validator.MinItems)),
	)

	if validator.MaxItems != nil {
		rule.Range[1] = rules.NewRuleLit(
			[]byte(fmt.Sprintf("%d", *validator.MaxItems)),
		)
	}

	return string(rule.Bytes())
}

func (validator *SliceValidator) PostValidate(rv reflect.Value) error {
	lenOfValue := uint64(0)
	if !rv.IsNil() {
		lenOfValue = uint64(rv.Len())
	}

	if lenOfValue < validator.MinItems {
		return &validatorerrors.OutOfRangeError{
			Topic:   "array items",
			Minimum: validator.MinItems,
			Current: lenOfValue,
		}
	}

	if validator.MaxItems != nil && lenOfValue > *validator.MaxItems {
		return &validatorerrors.OutOfRangeError{
			Topic:   "array items",
			Current: lenOfValue,
			Maximum: validator.MaxItems,
		}
	}
	return nil
}
