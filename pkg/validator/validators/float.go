package validators

import (
	"bytes"
	"fmt"
	"math"
	"strconv"
	"sync"

	"github.com/go-json-experiment/json/jsontext"
	validatorerrors "github.com/octohelm/courier/pkg/validator/errors"
	"github.com/octohelm/courier/pkg/validator/internal"
	"github.com/octohelm/courier/pkg/validator/internal/rules"
	"github.com/octohelm/x/ptr"
)

func init() {
	internal.Register(&floatValidatorProvider{})
}

type floatValidatorProvider struct{}

func (c *floatValidatorProvider) Validator(rule *rules.Rule) (internal.Validator, error) {
	validator := &FloatValidator{}

	switch rule.Name {
	case "float", "float32":
		validator.BitSize = 32
	case "double", "float64":
		validator.BitSize = 64
	}

	if rule.Params != nil {
		if len(rule.Params) > 2 {
			return nil, fmt.Errorf("float should only 1 or 2 parameter, but got %d", len(rule.Params))
		}

		maxDigitsBytes := rule.Params[0].Bytes()
		if len(maxDigitsBytes) > 0 {
			maxDigits, err := strconv.ParseUint(string(maxDigitsBytes), 10, 4)
			if err != nil {
				return nil, rules.NewSyntaxError("decimal digits should be a uint value which less than 16, but got `%s`", maxDigitsBytes)
			}
			validator.MaxDigits = ptr.Ptr(uint(maxDigits))
		}

		if len(rule.Params) > 1 {
			decimalDigitsBytes := rule.Params[1].Bytes()

			if len(decimalDigitsBytes) > 0 {
				decimalDigits, err := strconv.ParseUint(string(decimalDigitsBytes), 10, 4)
				if err != nil {
					return nil, rules.NewSyntaxError("decimal digits should be a uint value, but got `%s`", decimalDigitsBytes)
				}

				if validator.MaxDigits != nil && uint(decimalDigits) >= *validator.MaxDigits {
					return nil, rules.NewSyntaxError("decimal digits should be less than %d, but got `%s`", *validator.MaxDigits, decimalDigitsBytes)
				}

				validator.DecimalDigits = ptr.Ptr(uint(decimalDigits))
			}
		}
	}

	if err := validator.unmarshalRule(rule); err != nil {
		return nil, err
	}

	return validator, nil
}

/*
Validator for float32 and float64

Rules:

ranges

	@float[min,max]
	@float[1,10] // value should large or equal than 1 and less or equal than 10
	@float(1,10] // value should large than 1 and less or equal than 10
	@float[1,10) // value should large or equal than 1

	@float[1,)  // value should large or equal than 1
	@float[,1)  // value should less than 1

enumeration

	@float{1.1,1.2,1.3} // value should be one of these

multiple of some float value

	@float{%multipleOf}
	@float{%2.2} // value should be multiple of 2.2

max digits and decimal digits.
when defined, all values in rule should be under range of them.

	@float<MAX_DIGITS,DECIMAL_DIGITS>
	@float<5,2> // will checkout these values invalid: 1.111 (decimal digits too many), 12345.6 (digits too many)

composes

	@float<MAX_DIGITS,DECIMAL_DIGITS>[min,max]

aliases:

	@float32 = @float<7>
	@float64 = @float<15>
*/
type FloatValidator struct {
	BitSize       int
	MaxDigits     *uint
	DecimalDigits *uint

	Number[float64]
}

func (floatValidatorProvider) Names() []string {
	return []string{
		"float",
		"double",
		"float32",
		"float64",
	}
}

func (validator *FloatValidator) Validate(value jsontext.Value) error {
	if value.Kind() != '0' {
		return &validatorerrors.ErrInvalidType{
			Type:  "number",
			Value: string(value),
		}
	}

	val, err := strconv.ParseFloat(string(value), 64)
	if err != nil {
		return fmt.Errorf("invalid value %w", err)
	}

	if validator.Enums != nil {
		enums := make([]any, len(validator.Enums))
		in := false

		for _, enum := range validator.Enums {
			if enum == val {
				in = true
				break
			}
		}

		if !in {
			return &validatorerrors.NotInEnumError{
				Topic:   "float value",
				Current: val,
				Enums:   enums,
			}
		}

		return nil
	}

	if validator.Minimum != nil {
		mininum := *validator.Minimum
		if (validator.ExclusiveMinimum && val == mininum) || val < mininum {
			return &validatorerrors.OutOfRangeError{
				Topic:            "float value",
				Current:          val,
				Minimum:          mininum,
				ExclusiveMinimum: validator.ExclusiveMinimum,
			}
		}
	}

	if validator.Maximum != nil {
		maxinum := *validator.Maximum

		if (validator.ExclusiveMaximum && val == maxinum) || val > maxinum {
			return &validatorerrors.OutOfRangeError{
				Topic:            "float value",
				Current:          val,
				Maximum:          maxinum,
				ExclusiveMaximum: validator.ExclusiveMaximum,
			}
		}
	}

	get := sync.OnceValues(func() (uint, uint) {
		return lengthOfDigits(value)
	})

	if validator.MultipleOf != 0 {
		_, d := get()

		if !multipleOf(val, validator.MultipleOf, d) {
			return &validatorerrors.ErrMultipleOf{
				Topic:      "float value",
				Current:    val,
				MultipleOf: validator.MultipleOf,
			}
		}
	}

	if validator.DecimalDigits != nil {
		_, d := get()
		m := *validator.DecimalDigits
		if d > m {
			return &validatorerrors.OutOfRangeError{
				Topic:   "decimal digits of float value",
				Current: d,
				Maximum: m,
			}
		}
	}

	if validator.MaxDigits != nil {
		n, _ := get()
		m := *validator.MaxDigits
		if n > m {
			return &validatorerrors.OutOfRangeError{
				Topic:   "total digits of float value",
				Current: n,
				Maximum: m,
			}
		}
		return nil
	}

	return nil
}

func lengthOfDigits(value jsontext.Value) (uint, uint) {
	var n, d int

	nd := bytes.Split(value, []byte("."))
	n = len(nd[0])

	if len(nd) == 2 {
		d = len(nd[1])
	}

	if bytes.Equal(nd[0], []byte("-0")) {
		n = 0
	}

	return uint(n + d), uint(d)
}

func multipleOf(v float64, div float64, decimalDigits uint) bool {
	val := fixDecimal(v/div, int(decimalDigits))
	return val == math.Trunc(val)
}

func fixDecimal(f float64, n int) float64 {
	res, _ := strconv.ParseFloat(strconv.FormatFloat(f, 'g', n, 64), 64)
	return res
}

func (validator *FloatValidator) String() string {
	rule := rules.NewRule("float")

	if validator.MaxDigits != nil {
		if validator.DecimalDigits != nil {
			rule.Params = []rules.RuleNode{
				rules.NewRuleLit([]byte(strconv.Itoa(int(*validator.MaxDigits)))),
				rules.NewRuleLit([]byte(strconv.Itoa(int(*validator.DecimalDigits)))),
			}
		} else {
			rule.Params = []rules.RuleNode{
				rules.NewRuleLit([]byte(strconv.Itoa(int(*validator.MaxDigits)))),
			}
		}
	}

	validator.marshalRule(rule)

	return string(rule.Bytes())
}
