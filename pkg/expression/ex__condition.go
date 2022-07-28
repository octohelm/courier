package expression

import (
	"context"

	"github.com/octohelm/courier/pkg/expression/raw"
)

// Not like `!`
//
// Syntax
//
//	not(gt(1))
var Not = Register(func(b ExprBuilder) func(ex ...Expression) Expression {
	return func(rules ...Expression) Expression {
		args := make([]any, len(rules))
		for i := range rules {
			args[i] = rules[i]
		}
		return b.BuildExpression(args...)
	}
}, &not{})

type not struct {
	E
	Expr Expr
}

func (e *not) Exec(ctx context.Context, in any) (any, error) {
	o, err := e.Expr.Exec(ctx, in)
	if err != nil {
		return nil, err
	}
	return !raw.ToBool(raw.ValueOf(o)), nil
}

// AllOf all rules must match. like `&&`
//
// Syntax
//
//	allOf(gt(1), lt(10))
var AllOf = Register(func(b ExprBuilder) func(rules ...Expression) Expression {
	return func(rules ...Expression) Expression {
		args := make([]any, len(rules))
		for i := range rules {
			args[i] = rules[i]
		}
		return b.BuildExpression(args...)
	}
}, &allOf{})

type allOf struct {
	E
	rules []Expr
}

func (e *allOf) InitFromArgs(args []any) error {
	e.rules = make([]Expr, 0, len(args))
	for i := range args {
		switch x := args[i].(type) {
		case Expr:
			e.rules = append(e.rules, x)
		default:
		}
	}
	return nil
}

func (e *allOf) Exec(ctx context.Context, in any) (any, error) {
	for i := range e.rules {
		o, err := e.rules[i].Exec(ctx, in)
		if err != nil {
			return nil, err
		}
		if !raw.ToBool(raw.ValueOf(o)) {
			return false, nil
		}
	}
	return true, nil
}

// AnyOf at least one rule matched. like `||`
//
// Syntax
//
//	anyOf(eq(1), lt(10))
var AnyOf = Register(func(b ExprBuilder) func(rules ...Expression) Expression {
	return func(rules ...Expression) Expression {
		args := make([]any, len(rules))
		for i := range rules {
			args[i] = rules[i]
		}
		return b.BuildExpression(args...)
	}
}, &anyOf{})

type anyOf struct {
	E
	rules []Expr
}

func (e *anyOf) InitFromArgs(args []any) error {
	e.rules = make([]Expr, 0, len(args))
	for i := range args {
		switch x := args[i].(type) {
		case Expr:
			e.rules = append(e.rules, x)
		default:
		}
	}
	return nil
}

func (e *anyOf) Exec(ctx context.Context, in any) (any, error) {
	for i := range e.rules {
		o, err := e.rules[i].Exec(ctx, in)
		if err != nil {
			return nil, err
		}
		if raw.ToBool(raw.ValueOf(o)) {
			return true, nil
		}
	}
	return false, nil
}

// OneOf only one rule match.
//
// Syntax
//
//	oneOf(1, 2, 3)
var OneOf = Register(func(b ExprBuilder) func(ex ...any) Expression {
	return func(ex ...any) Expression {
		return b.BuildExpression(ex...)
	}
}, &oneOf{})

type oneOf struct {
	E
	rules []Expr
}

func (e *oneOf) InitFromArgs(args []any) error {
	list := make([]Expr, len(args))
	for i := range args {
		switch x := args[i].(type) {
		case Expr:
			list[i] = x
		default:
			list[i] = &eq{
				Expect: raw.ValueOf(x),
			}
		}
	}
	e.rules = list
	return nil
}

func (e *oneOf) Exec(ctx context.Context, in any) (any, error) {
	matched := 0

	for i := range e.rules {
		o, err := e.rules[i].Exec(ctx, in)
		if err != nil {
			return nil, err
		}

		if raw.ToBool(raw.ValueOf(o)) {
			matched++
		}

		if matched > 1 {
			return false, nil
		}
	}

	return matched == 1, nil
}
