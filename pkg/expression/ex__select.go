package expression

import (
	"context"
)

// Select
//
// Syntax
//
//	select(
//		when(
//			pipe(get("x"), eq(1)),
//			eq(2),
//	  	),
//	  	eq(1),
//	)
var Select = Register(func(b ExprBuilder) func(rules ...Expression) Expression {
	return func(rules ...Expression) Expression {
		args := make([]any, len(rules))
		for i := range rules {
			args[i] = rules[i]
		}
		return b.BuildExpression(args...)
	}
}, &selectExpr{})

type selectExpr struct {
	E     `name:"select"`
	Rules []any `arg:"..."`
}

func (e *selectExpr) Exec(ctx context.Context, in any) (any, error) {
	for i := range e.Rules {
		switch x := e.Rules[i].(type) {
		case Expr:
			o, err := x.Exec(ctx, in)
			if err != nil {
				return nil, err
			}
			if o != nil {
				return o, nil
			}
		default:
			if x != nil {
				return x, nil
			}
		}
	}
	return nil, nil
}
