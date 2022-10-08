package expression

import (
	"context"
	"github.com/octohelm/courier/pkg/expression/raw"
)

// Not
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

// AllOf
//
// Syntax
//
//	allOf(gt(1), lt(10))
var AllOf = Register(func(b ExprBuilder) func(ex ...Expression) Expression {
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
	Rules []any `arg:"..."`
}

func (e *allOf) Exec(ctx context.Context, in any) (any, error) {
	for i := range e.Rules {
		switch x := e.Rules[i].(type) {
		case Expr:
			o, err := x.Exec(ctx, in)
			if err != nil {
				return nil, err
			}
			if !raw.ToBool(raw.ValueOf(o)) {
				return false, nil
			}
		default:
		}
	}

	return true, nil
}

// OneOf
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
	Rules []any `arg:"..."`
}

func (e *oneOf) Exec(ctx context.Context, in any) (any, error) {
	for i := range e.Rules {
		switch x := e.Rules[i].(type) {
		case Expr:
			o, err := x.Exec(ctx, in)
			if err != nil {
				return nil, err
			}
			if raw.ToBool(raw.ValueOf(o)) {
				return true, nil
			}
		default:
			ret, err := raw.Compare(raw.ValueOf(x), raw.ValueOf(in))
			if err != nil {
				return false, err
			}
			if ret == 0 {
				return true, nil
			}
		}
	}

	return false, nil
}
