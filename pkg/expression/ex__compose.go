package expression

import (
	"context"
	"github.com/octohelm/courier/pkg/expression/raw"
	"github.com/pkg/errors"
)

// Each
//
// Syntax
//
//		each(
//	   key(len(), gt(1)),
//	   elem(len(), gt(1)),
//		)
var Each = Register(func(b ExprBuilder) func(ex Expression) Expression {
	return func(ex Expression) Expression {
		return b.BuildExpression(ex)
	}
}, &each{})

type each struct {
	E
	Rules []any `arg:"..."`
}

func (e *each) Exec(ctx context.Context, in any) (any, error) {
	switch x := raw.ValueOf(in).(type) {
	case raw.MapValue:
		iter := x.Iter()

		for iter.Next() {
			val := iter.Val()

			for j := range e.Rules {
				switch r := e.Rules[j].(type) {
				case Expr:
					ret, err := r.Exec(ctx, val)
					if err != nil {
						return nil, err
					}
					if !raw.ToBool(raw.ValueOf(ret)) {
						return false, nil
					}
				}
			}
		}
	case raw.ArrayValue:
		iter := x.Iter()

		for iter.Next() {
			val := iter.Val()

			for j := range e.Rules {
				switch r := e.Rules[j].(type) {
				case Expr:
					ret, err := r.Exec(ctx, val)
					if err != nil {
						return nil, err
					}
					if !raw.ToBool(raw.ValueOf(ret)) {
						return false, nil
					}
				}
			}
		}
	}
	return true, nil
}

// Elem
//
// Syntax
//
//	elem(eq(1))
var Elem = Register(func(b ExprBuilder) func(ex Expression) Expression {
	return func(ex Expression) Expression {
		return b.BuildExpression(ex)
	}
}, &elem{})

type elem struct {
	E
	Elem Expr
}

func (e *elem) Exec(ctx context.Context, in any) (any, error) {
	switch x := in.(type) {
	case raw.IterVal:
		return e.Elem.Exec(ctx, x.Value())
	}
	return nil, errors.New("`elem` must be used in `each`")
}

// Key
//
// Syntax
//
//	key(eq(1))
var Key = Register(func(b ExprBuilder) func(ex Expression) Expression {
	return func(ex Expression) Expression {
		return b.BuildExpression(ex)
	}
}, &key{})

type key struct {
	E
	Key Expr
}

func (e *key) Exec(ctx context.Context, in any) (any, error) {
	switch x := in.(type) {
	case raw.IterVal:
		return e.Key.Exec(ctx, x.Key())
	}
	return nil, errors.New("`key` must be used in `each`")
}
