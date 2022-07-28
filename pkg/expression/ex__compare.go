package expression

import (
	"context"

	"github.com/octohelm/courier/pkg/expression/raw"
)

// Eq
//
// Syntax
//
//	eq(1)
//	eq(1.0)
//	eq("x")
var Eq = Register(func(b ExprBuilder) func(expect any) Expression {
	return func(expect any) Expression {
		return b.BuildExpression(expect)
	}
}, &eq{})

type eq struct {
	E
	Expect raw.Value
}

func (e *eq) Exec(ctx context.Context, in any) (any, error) {
	ret, err := raw.Compare(raw.ValueOf(in), e.Expect)
	if err != nil {
		return nil, err
	}
	return ret == 0, nil
}

// Lt
//
// Syntax
//
//	lt(10)
var Lt = Register(func(b ExprBuilder) func(expect any) Expression {
	return func(expect any) Expression {
		return b.BuildExpression(expect)
	}
}, &lt{})

type lt struct {
	E
	Max raw.Value
}

func (e *lt) Exec(ctx context.Context, in any) (any, error) {
	ret, err := raw.Compare(raw.ValueOf(in), e.Max)
	if err != nil {
		return nil, err
	}
	return ret < 0, nil
}

// Lte
//
// Syntax
//
//	lte(10)
var Lte = Register(func(b ExprBuilder) func(expect any) Expression {
	return func(expect any) Expression {
		return b.BuildExpression(expect)
	}
}, &lte{})

type lte struct {
	E
	Max raw.Value
}

func (e *lte) Exec(ctx context.Context, in any) (any, error) {
	ret, err := raw.Compare(raw.ValueOf(in), e.Max)
	if err != nil {
		return nil, err
	}
	return ret <= 0, nil
}

// Gt
//
// Syntax
//
//	gt(10)
var Gt = Register(func(b ExprBuilder) func(expect any) Expression {
	return func(expect any) Expression {
		return b.BuildExpression(expect)
	}
}, &gt{})

type gt struct {
	E
	Min raw.Value
}

func (e *gt) Exec(ctx context.Context, in any) (any, error) {
	ret, err := raw.Compare(raw.ValueOf(in), e.Min)
	if err != nil {
		return nil, err
	}
	return ret > 0, nil
}

// Gte
//
// Syntax
//
//	gte(10)
var Gte = Register(func(b ExprBuilder) func(expect any) Expression {
	return func(expect any) Expression {
		return b.BuildExpression(expect)
	}
}, &gte{})

type gte struct {
	E
	Min raw.Value
}

func (e *gte) Exec(ctx context.Context, in any) (any, error) {
	ret, err := raw.Compare(raw.ValueOf(in), e.Min)
	if err != nil {
		return nil, err
	}
	return ret >= 0, nil
}
