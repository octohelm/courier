package expression

import (
	"context"
	"errors"
	"fmt"

	contextx "github.com/octohelm/x/context"
)

// Get
//
// Syntax
//
//	get("x")
var Get = Register(func(b ExprBuilder) func(fieldName string) Expression {
	return func(fieldName string) Expression {
		return b.BuildExpression(fieldName)
	}
}, &get{})

type get struct {
	E
	FieldName string
}

func (e *get) Exec(ctx context.Context, in any) (any, error) {
	vg := ValueGetterFromContext(ctx)
	if vg == nil {
		return nil, errors.New("missing value getter")
	}
	if v, ok := vg.Get(e.FieldName); ok {
		return v, nil
	}
	return nil, fmt.Errorf("missing field %s", e.FieldName)
}

type ValueGetter interface {
	Get(name string) (any, bool)
}

type contextValueGetter struct{}

func ValueGetterFunc(g func(name string) (any, bool)) ValueGetter {
	return &valueGetterFunc{fn: g}
}

type valueGetterFunc struct {
	fn func(name string) (any, bool)
}

func (v *valueGetterFunc) Get(name string) (any, bool) {
	return v.fn(name)
}

func WithValueGetter(ctx context.Context, getter ValueGetter) context.Context {
	return contextx.WithValue(ctx, contextValueGetter{}, getter)
}

func ValueGetterFromContext(ctx context.Context) ValueGetter {
	if vg, ok := ctx.Value(contextValueGetter{}).(ValueGetter); ok {
		return vg
	}
	return nil
}
