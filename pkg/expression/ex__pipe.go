package expression

import (
	"context"
)

// Pipe
//
// Syntax
//
//	pipe(1, eq(1))
var Pipe = Register(func(b ExprBuilder) func(ex ...any) Expression {
	return func(ex ...any) Expression {
		return b.BuildExpression(ex...)
	}
}, &pipe{})

type pipe struct {
	E
	Expressions []any `arg:"..."`
}

func (e *pipe) Exec(ctx context.Context, in any) (any, error) {
	out := in

	for i := range e.Expressions {
		switch x := e.Expressions[i].(type) {
		case Expr:
			o, err := x.Exec(ctx, out)
			if err != nil {
				return nil, err
			}
			out = o
		default:
			out = x
		}
	}

	return out, nil
}
