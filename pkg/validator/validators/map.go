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
	internal.Register(&mapValidatorProvider{})
}

type mapValidatorProvider struct{}

func (mapValidatorProvider) Names() []string {
	return []string{"map", "record"}
}

func (c *mapValidatorProvider) Validator(rule *rules.Rule) (internal.Validator, error) {
	validator := &MapValidator{}

	if rule.ExclusiveLeft || rule.ExclusiveRight {
		return nil, rules.NewSyntaxError("range mark of %s should not be `(` or `)`", c.Names()[0])
	}

	if rule.Params != nil {
		if len(rule.Params) != 2 {
			return nil, fmt.Errorf("map should only 2 parameter, but got %d", len(rule.Params))
		}

		for i, param := range rule.Params {
			switch r := param.(type) {
			case *rules.Rule:
				switch i {
				case 0:
					validator.key.Rule = string(r.RAW)
				case 1:
					validator.elem.Rule = string(r.RAW)
				}
			case *rules.RuleLit:
				validator.elem.Rule = string(r.Bytes())
			}
		}
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
			validator.MinProperties = *minn
		}

		validator.MaxProperties = maxn
	}

	return validator, nil
}

/*
Validator for map

Rules:

	@map<KEY_RULE, ELEM_RULE>[minSize,maxSize]
	@map<KEY_RULE, ELEM_RULE>[length]

	@map<@string{A,B,C},@int[0]>[,100]
*/
type MapValidator struct {
	key  internal.ValidatorOption
	elem internal.ValidatorOption

	MinProperties uint64
	MaxProperties *uint64
}

func (validator *MapValidator) Key() internal.ValidatorOption {
	return validator.key
}

func (validator *MapValidator) Elem() internal.ValidatorOption {
	return validator.elem
}

func (validator *MapValidator) Validate(value jsontext.Value) error {
	return nil
}

func (validator *MapValidator) String() string {
	rule := rules.NewRule("map")

	if validator.key.Rule != "" || validator.elem.Rule != "" {
		rule.Params = []rules.RuleNode{
			rules.NewRuleLit([]byte(validator.key.Rule)),
			rules.NewRuleLit([]byte(validator.elem.Rule)),
		}
	}

	rule.Range = make([]*rules.RuleLit, 2)

	rule.Range[0] = rules.NewRuleLit(
		[]byte(fmt.Sprintf("%d", validator.MinProperties)),
	)

	if validator.MaxProperties != nil {
		rule.Range[1] = rules.NewRuleLit(
			[]byte(fmt.Sprintf("%d", *validator.MaxProperties)),
		)
	}

	return string(rule.Bytes())
}

func (validator *MapValidator) PostValidate(rv reflect.Value) error {
	lenOfValue := uint64(0)
	if !rv.IsNil() {
		lenOfValue = uint64(rv.Len())
	}

	if lenOfValue < validator.MinProperties {
		return &validatorerrors.OutOfRangeError{
			Topic:   "props count",
			Minimum: validator.MinProperties,
			Current: lenOfValue,
		}
	}

	if validator.MaxProperties != nil && lenOfValue > *validator.MaxProperties {
		return &validatorerrors.OutOfRangeError{
			Topic:   "props count",
			Current: lenOfValue,
			Maximum: validator.MaxProperties,
		}
	}

	return nil
}
