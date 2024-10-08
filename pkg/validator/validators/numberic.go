package validators

import (
	"fmt"

	"github.com/go-json-experiment/json"
	"github.com/octohelm/courier/pkg/validator/internal/rules"
)

type numeric[T ~int | ~int8 | ~int16 | ~int32 | ~int64 | ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~float32 | ~float64] struct {
	Minimum          *T
	Maximum          *T
	MultipleOf       T
	ExclusiveMaximum bool
	ExclusiveMinimum bool
	Enums            []T
}

func (n *numeric[T]) unmarshalRule(rule *rules.Rule) error {
	if rule.Range != nil {
		switch len(rule.Range) {
		case 2:
			if value := rule.Range[0].Bytes(); len(value) != 0 {
				n.Minimum = new(T)
				if err := json.Unmarshal(value, n.Minimum); err != nil {
					return fmt.Errorf("invalid min value %v", string(value))
				}
			}
			if value := rule.Range[1].Bytes(); len(value) != 0 {
				n.Maximum = new(T)
				if err := json.Unmarshal(value, n.Maximum); err != nil {
					return fmt.Errorf("invalid max value %v", string(value))
				}
			}
		case 1:
			if value := rule.Range[0].Bytes(); len(value) != 0 {
				n.Minimum = new(T)
				if err := json.Unmarshal(value, n.Minimum); err != nil {
					return fmt.Errorf("invalid min value %v", string(value))
				}
			}
			n.Maximum = n.Minimum
		}
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

			for _, v := range ruleValues {
				var enum T
				if err := json.Unmarshal(v.Bytes(), &enum); err != nil {
					return fmt.Errorf("invalid enum value %v", string(v.Bytes()))
				}
				n.Enums = append(n.Enums, enum)
			}
		}
	}
	return nil
}

func (n *numeric[T]) marshalRule(rule *rules.Rule) {
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
	} else if n.Enums != nil {
		ruleValues := make([]*rules.RuleLit, 0)
		for _, v := range n.Enums {
			ruleValues = append(ruleValues, rules.NewRuleLit([]byte(fmt.Sprintf("%v", v))))
		}
		rule.ValueMatrix = [][]*rules.RuleLit{ruleValues}
	}
}
