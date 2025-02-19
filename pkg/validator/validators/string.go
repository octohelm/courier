package validators

import (
	"fmt"
	"regexp"
	"unicode/utf8"

	validatorerrors "github.com/octohelm/courier/pkg/validator/errors"

	"github.com/go-json-experiment/json/jsontext"
	"github.com/octohelm/courier/pkg/validator/internal"
	"github.com/octohelm/courier/pkg/validator/internal/rules"
)

func init() {
	internal.Register(&stringValidatorProvider{})
}

type stringValidatorProvider struct{}

func (stringValidatorProvider) Names() []string {
	return []string{
		"string", "char",
	}
}

func (c *stringValidatorProvider) Validator(rule *rules.Rule) (internal.Validator, error) {
	validator := &StringValidator{
		LenMode: StrLenModeLength,
	}

	if rule.ExclusiveLeft || rule.ExclusiveRight {
		return nil, rules.NewSyntaxError("range mark of %s should not be `(` or `)`", c.Names()[0])
	}

	if rule.Params != nil {
		if len(rule.Params) != 1 {
			return nil, fmt.Errorf("string should only 1 parameter, but got %d", len(rule.Params))
		}
		lenMode := StrLenMode(rule.Params[0].Bytes())
		if lenMode != StrLenModeLength && lenMode != StrLenModeRuneCount {
			return nil, fmt.Errorf("invalid len mode %s", lenMode)
		}
		if lenMode != StrLenModeLength {
			validator.LenMode = lenMode
		}
	} else if rule.Name == "char" {
		validator.LenMode = StrLenModeRuneCount
	}

	if rule.Pattern != "" {
		validator.Pattern = regexp.MustCompile(rule.Pattern).String()
		return validator, nil
	}

	ruleValues := rule.ComputedValues()

	if ruleValues != nil {
		for _, v := range ruleValues {
			validator.Enums = append(validator.Enums, string(v.Bytes()))
		}
	}

	if rule.Range != nil {
		minn, maxn, err := convertRangeValues[uint64](rule)
		if err != nil {
			return nil, err
		}
		if minn != nil {
			validator.MinLength = *minn
		}
		validator.MaxLength = maxn
	}

	return validator, nil
}

/*
Validator for string

Rules:

	@string/regexp/
	@string{VALUE_1,VALUE_2,VALUE_3}
	@string<StrLenMode>[from,to]
	@string<StrLenMode>[length]

ranges

	@string[min,max]
	@string[length]
	@string[1,10] // string length should large or equal than 1 and less or equal than 10
	@string[1,]  // string length should large or equal than 1 and less than the maxinum of int32
	@string[,1]  // string length should less than 1 and large or equal than 0
	@string[10]  // string length should be equal 10

enumeration

	@string{A,B,C} // should one of these values

regexp

	@string/\w+/ // string values should match \w+

since we use / as wrapper for regexp, we need to use \ to escape /

length mode in parameter

	@string<length> // use string length directly
	@string<rune_count> // use rune count as string length

composes

	@string<>[1,]

aliases:

	@char = @string<rune_count>
*/
type StringValidator struct {
	Pattern   string
	LenMode   StrLenMode
	MinLength uint64
	MaxLength *uint64
	Enums     []string
}

type StrLenMode string

const (
	StrLenModeLength    StrLenMode = "length"
	StrLenModeRuneCount StrLenMode = "rune_count"
)

var strLenModes = map[StrLenMode]func(s string) uint64{
	StrLenModeLength: func(s string) uint64 {
		return uint64(len(s))
	},
	StrLenModeRuneCount: func(s string) uint64 {
		return uint64(utf8.RuneCount([]byte(s)))
	},
}

func (validator *StringValidator) Validate(value jsontext.Value) error {
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

	if validator.Enums != nil {
		enums := make([]any, len(validator.Enums))
		in := false

		for i := range validator.Enums {
			enums[i] = validator.Enums[i]

			if validator.Enums[i] == val {
				in = true
				break
			}
		}

		if !in {
			return &validatorerrors.NotInEnumError{
				Topic:   "string value",
				Current: val,
				Enums:   enums,
			}
		}

		return nil
	}

	if validator.Pattern != "" {
		matched, _ := regexp.MatchString(validator.Pattern, val)
		if !matched {
			return &validatorerrors.ErrNotMatch{
				Topic:   "string value",
				Pattern: validator.Pattern,
				Current: val,
			}
		}
		return nil
	}

	strLen := strLenModes[validator.LenMode](val)

	if strLen < validator.MinLength {
		return &validatorerrors.OutOfRangeError{
			Topic:   "string value length",
			Current: strLen,
			Minimum: validator.MinLength,
		}
	}

	if validator.MaxLength != nil && strLen > *validator.MaxLength {
		return &validatorerrors.OutOfRangeError{
			Topic:   "string value length",
			Current: strLen,
			Maximum: *validator.MaxLength,
		}
	}
	return nil
}

func (validator *StringValidator) String() string {
	rule := rules.NewRule("string")

	rule.Params = []rules.RuleNode{
		rules.NewRuleLit([]byte(validator.LenMode)),
	}

	if validator.Enums != nil {
		ruleValues := make([]*rules.RuleLit, 0)
		for _, e := range validator.Enums {
			ruleValues = append(ruleValues, rules.NewRuleLit([]byte(e)))
		}
		rule.ValueMatrix = [][]*rules.RuleLit{ruleValues}

		return string(rule.Bytes())
	}

	if validator.Pattern != "" {
		rule.Pattern = validator.Pattern

		return string(rule.Bytes())
	}

	rule.Range = make([]*rules.RuleLit, 2)

	rule.Range[0] = rules.NewRuleLit(
		[]byte(fmt.Sprintf("%d", validator.MinLength)),
	)

	if validator.MaxLength != nil {
		rule.Range[1] = rules.NewRuleLit(
			[]byte(fmt.Sprintf("%d", *validator.MaxLength)),
		)
	}

	return string(rule.Bytes())
}
