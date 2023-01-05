package expression

import (
	"context"
	"unicode/utf8"

	"github.com/octohelm/courier/pkg/expression/raw"
)

// Len
//
// Syntax
//
//	len()
var Len = Register(func(b ExprBuilder) func() Expression {
	return func() Expression {
		return b.BuildExpression()
	}
}, &lenExpr{})

type lenExpr struct {
	E `name:"len"`
}

func (e *lenExpr) Exec(ctx context.Context, in any) (any, error) {
	return raw.Len(raw.ValueOf(in)), nil
}

// CharCount
//
// Syntax
//
//	charCount()
var CharCount = Register(func(b ExprBuilder) func() Expression {
	return func() Expression {
		return b.BuildExpression()
	}
}, &charCount{})

type charCount struct {
	E
}

func (e *charCount) Exec(ctx context.Context, in any) (any, error) {
	if s, ok := (raw.ValueOf(in)).(raw.StringValue); ok {
		return utf8.RuneCount([]byte(s)), nil
	}
	return 0, nil
}
