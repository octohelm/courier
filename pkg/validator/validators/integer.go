package validators

import (
	"bytes"
	"fmt"
	"strconv"
	"unicode"

	"github.com/octohelm/x/ptr"

	"github.com/go-json-experiment/json/jsontext"
	validatorerrors "github.com/octohelm/courier/pkg/validator/errors"
	"github.com/octohelm/courier/pkg/validator/internal"
	"github.com/octohelm/courier/pkg/validator/internal/rules"
)

func init() {
	internal.Register(&integerValidatorProvider{})
}

type integerValidatorProvider struct {
}

func (integerValidatorProvider) Names() []string {
	return []string{
		"int", "int8", "int16", "int32", "int64",
		"uint", "uint8", "uint16", "uint32", "uint64",
	}
}

func (c *integerValidatorProvider) Validator(rule *rules.Rule) (internal.Validator, error) {
	bitSize, unsigned, err := c.readBitSize(rule)
	if err != nil {
		return nil, err
	}

	if unsigned {
		validator := &IntegerValidator[uint64]{
			Unsigned: true,
			BitSize:  bitSize,
		}

		if err := validator.unmarshalRule(rule); err != nil {
			return nil, err
		}

		validator.SetDefaults()

		return validator, nil
	}

	validator := &IntegerValidator[int64]{
		BitSize: bitSize,
	}

	if err := validator.unmarshalRule(rule); err != nil {
		return nil, err
	}

	validator.SetDefaults()

	return validator, nil
}

func (c *integerValidatorProvider) readBitSize(rule *rules.Rule) (uint, bool, error) {
	unsigned := false

	bitSizeBuf := &bytes.Buffer{}

	for i, char := range rule.Name {
		if i == 0 {
			if char == 'u' {
				unsigned = true
			}
		}

		if unicode.IsDigit(char) {
			bitSizeBuf.WriteRune(char)
		}
	}

	if bitSizeBuf.Len() == 0 && rule.Params != nil {
		if len(rule.Params) != 1 {
			return 0, false, fmt.Errorf("int should only 1 parameter, but got %d", len(rule.Params))
		}

		bitSizeBuf.Write(rule.Params[0].Bytes())
	}

	if bitSizeBuf.Len() != 0 {
		bitSizeStr := bitSizeBuf.String()
		bitSizeNum, err := strconv.ParseUint(bitSizeStr, 10, 8)
		if err != nil || bitSizeNum > 64 {
			return 0, false, rules.NewSyntaxError("(u)int parameter should be valid bit size, but got `%s`", bitSizeStr)
		}
		return uint(bitSizeNum), unsigned, nil
	}

	return 32, unsigned, nil
}

/*
Rules:

ranges

	@int[min,max]
	@int[1,10] // value should large or equal than 1 and less or equal than 10
	@int(1,10] // value should large than 1 and less or equal than 10
	@int[1,10) // value should large or equal than 1

	@int[1,)  // value should large or equal than 1 and less than the maxinum of int32
	@int[,1)  // value should less than 1 and large or equal than the mininum of int32
	@int  // value should less or equal than maxinum of int32 and large or equal than the mininum of int32

enumeration

	@int{1,2,3} // should one of these values

multiple of some integer value

	@int{%multipleOf}
	@int{%2} // should be multiple of 2

bit size in parameter

	@int<8>
	@int<16>
	@int<32>
	@int<64>

composes

	@int<8>[1,]

aliases:

	@int8 = @int<8>
	@int16 = @int<16>
	@int32 = @int<32>
	@int64 = @int<64>

Tips:
for JavaScript https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/Number/MAX_SAFE_INTEGER and https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/Number/MIN_SAFE_INTEGER

	int<53>
*/
type IntegerValidator[T ~int64 | ~uint64] struct {
	Unsigned bool
	BitSize  uint

	Number[T]
}

func (validator *IntegerValidator[T]) SetDefaults() {
	if validator.Unsigned {
		if validator.Minimum == nil {
			validator.Minimum = ptr.Ptr(T(0))
		}

		if validator.Maximum == nil {
			validator.Maximum = ptr.Ptr(T(internal.MaxUint(validator.BitSize)))
		}

		return
	}

	if validator.Minimum == nil {
		validator.Minimum = ptr.Ptr(T(internal.MinInt(validator.BitSize)))
	}

	if validator.Maximum == nil {
		validator.Maximum = ptr.Ptr(T(internal.MaxInt(validator.BitSize)))
	}
}

func (validator *IntegerValidator[T]) Validate(value jsontext.Value) error {
	if value.Kind() != '0' {
		return &validatorerrors.ErrInvalidType{
			Type:  "integer",
			Value: string(value),
		}
	}

	val := *new(T)

	switch any(val).(type) {
	case uint64:
		v, err := strconv.ParseUint(string(value), 10, 64)
		if err != nil {
			return fmt.Errorf("invalid integer value %w", err)
		}
		val = T(v)
	case int64:
		v, err := strconv.ParseInt(string(value), 10, 64)
		if err != nil {
			return fmt.Errorf("invalid integer value %w", err)
		}
		val = T(v)
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
				Topic:   "integer value",
				Current: val,
				Enums:   enums,
			}
		}

		return nil
	}

	mininum := *validator.Minimum
	maxinum := *validator.Maximum

	if ((validator.ExclusiveMinimum && val == mininum) || val < mininum) ||
		((validator.ExclusiveMaximum && val == maxinum) || val > maxinum) {
		return &validatorerrors.OutOfRangeError{
			Topic:            "integer value",
			Current:          val,
			Minimum:          mininum,
			ExclusiveMinimum: validator.ExclusiveMinimum,
			Maximum:          maxinum,
			ExclusiveMaximum: validator.ExclusiveMaximum,
		}
	}

	if validator.MultipleOf != 0 {
		if val%validator.MultipleOf != 0 {
			return &validatorerrors.ErrMultipleOf{
				Topic:      "integer value",
				Current:    val,
				MultipleOf: validator.MultipleOf,
			}
		}
	}

	return nil
}

func (validator *IntegerValidator[T]) String() string {
	name := "int"
	if validator.Unsigned {
		name = "uint"
	}
	rule := rules.NewRule(name)

	rule.Params = []rules.RuleNode{
		rules.NewRuleLit([]byte(strconv.Itoa(int(validator.BitSize)))),
	}

	validator.marshalRule(rule)

	return string(rule.Bytes())
}
