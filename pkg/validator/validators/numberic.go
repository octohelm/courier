package validators

import (
	"fmt"

	"github.com/go-json-experiment/json"
	"github.com/octohelm/courier/pkg/validator/internal/rules"
)

type Number[T ~int | ~int8 | ~int16 | ~int32 | ~int64 | ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~float32 | ~float64] struct {
	Minimum          *T
	Maximum          *T
	MultipleOf       T
	ExclusiveMaximum bool
	ExclusiveMinimum bool
	Enums            []T
}

func (n *Number[T]) unmarshalRule(rule *rules.Rule) error {
	if rule.Range != nil {
		minn, maxn, err := convertRangeValues[T](rule)
		if err != nil {
			return err
		}
		n.Minimum = minn
		n.Maximum = maxn
		n.ExclusiveMinimum = rule.ExclusiveLeft
		n.ExclusiveMaximum = rule.ExclusiveRight
	}

	if ruleValues := rule.ComputedValues(); ruleValues != nil {
		if len(ruleValues) == 1 {
			mayBeMultipleOf := ruleValues[0].Bytes()

			if mayBeMultipleOf[0] == '%' {
				value := mayBeMultipleOf[1:]
				if err := json.Unmarshal(value, &n.MultipleOf); err != nil {
					return fmt.Errorf("invalid multipleOf value %v", string(value))
				}
				return nil
			}
		}

		for _, v := range ruleValues {
			var enum T
			if err := json.Unmarshal(v.Bytes(), &enum); err != nil {
				return fmt.Errorf("invalid enum value %v", string(v.Bytes()))
			}
			n.Enums = append(n.Enums, enum)
		}
	}
	return nil
}

func convertRangeValues[T any](rule *rules.Rule) (min *T, max *T, err error) {
	switch len(rule.Range) {
	case 2:
		if value := rule.Range[0].Bytes(); len(value) != 0 {
			v := new(T)
			if err := json.Unmarshal(value, v); err != nil {
				return nil, nil, fmt.Errorf("invalid min value %v", string(value))
			}
			min = v
		}
		if value := rule.Range[1].Bytes(); len(value) != 0 {
			v := new(T)
			if err := json.Unmarshal(value, v); err != nil {
				return nil, nil, fmt.Errorf("invalid max value %v", string(value))
			}
			max = v
		}
	case 1:
		if value := rule.Range[0].Bytes(); len(value) != 0 {
			v := new(T)
			if err := json.Unmarshal(value, v); err != nil {
				return nil, nil, fmt.Errorf("invalid min value %v", string(value))
			}
			min = v
			max = v
		}
	}
	return min, max, err
}

func (n *Number[T]) marshalRule(rule *rules.Rule) {
	if n.Enums != nil {
		ruleValues := make([]*rules.RuleLit, 0)
		for _, v := range n.Enums {
			ruleValues = append(ruleValues, rules.NewRuleLit([]byte(fmt.Sprintf("%v", v))))
		}
		rule.ValueMatrix = [][]*rules.RuleLit{ruleValues}

		return
	}

	if n.Minimum != nil || n.Maximum != nil {
		rule.Range = make([]*rules.RuleLit, 2)

		if n.Minimum != nil {
			rule.Range[0] = rules.NewRuleLit(
				[]byte(fmt.Sprintf("%v", *n.Minimum)),
			)
		}

		if n.Maximum != nil {
			rule.Range[1] = rules.NewRuleLit(
				[]byte(fmt.Sprintf("%v", *n.Maximum)),
			)
		}

		rule.ExclusiveLeft = n.ExclusiveMinimum
		rule.ExclusiveRight = n.ExclusiveMaximum
	}

	rule.ExclusiveLeft = n.ExclusiveMinimum
	rule.ExclusiveRight = n.ExclusiveMaximum

	if n.MultipleOf != 0 {
		rule.ValueMatrix = [][]*rules.RuleLit{{
			rules.NewRuleLit([]byte("%" + fmt.Sprintf("%v", n.MultipleOf))),
		}}
	}
}
