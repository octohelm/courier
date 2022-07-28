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
//	   		key(len(), gt(1)),
//	   		elem(len(), gt(1)),
//		)
var Each = Register(func(b ExprBuilder) func(rules ...Expression) Expression {
	return func(rules ...Expression) Expression {
		args := make([]any, len(rules))
		for i := range rules {
			args[i] = rules[i]
		}

		return b.BuildExpression(args...)
	}
}, &each{})

type each struct {
	E
	Rules []any `arg:"..."`
}

func (e *each) Exec(ctx context.Context, in any) (any, error) {
	switch x := raw.ValueOf(in).(type) {
	case raw.Iterable:
		c, cancel := context.WithCancel(context.Background())
		defer cancel()
		for item := range x.Iter(c) {
			for j := range e.Rules {
				switch r := e.Rules[j].(type) {
				case Expr:
					ret, err := r.Exec(ctx, item)
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
	case raw.Entity:
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
	case raw.Entity:
		return e.Key.Exec(ctx, x.Key())
	}
	return nil, errors.New("`key` must be used in `each`")
}
