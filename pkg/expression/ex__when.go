package expression

import (
	"context"

	"github.com/octohelm/courier/pkg/expression/raw"
)

// When
//
// Syntax
//
//	when(
//		pipe(get("x"), eq(1)),
//		eq(2),
//	)
var When = Register(func(b ExprBuilder) func(condition Expression, then Expression) Expression {
	return func(condition Expression, then Expression) Expression {
		return b.BuildExpression(condition, then)
	}
}, &when{})

type when struct {
	E
	Condition Expr
	Then      Expr
}

func (e *when) Exec(ctx context.Context, in any) (any, error) {
	condRet, err := e.Condition.Exec(ctx, in)
	if err != nil {
		return nil, err
	}
	if raw.ToBool(raw.ValueOf(condRet)) {
		return e.Then.Exec(ctx, in)
	}
	return nil, nil
}
