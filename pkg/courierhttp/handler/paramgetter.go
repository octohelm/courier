package handler

import "context"

type ParamGetter interface {
	ByName(k string) string
}

type paramGetterCtx struct{}

func ParamGetterFromContext(ctx context.Context) ParamGetter {
	if g, ok := ctx.Value(paramGetterCtx{}).(ParamGetter); ok {
		return g
	}
	return Params{}
}

func ContextWithParamGetter(ctx context.Context, p ParamGetter) context.Context {
	return context.WithValue(ctx, paramGetterCtx{}, p)
}

type Params map[string]string

func (d Params) ByName(k string) string {
	v, _ := d[k]
	return v
}
