package httprequest

import (
	"context"

	contextx "github.com/octohelm/x/context"
)

type PathValueGetter interface {
	PathValue(k string) string
}

var paramGetterContext = contextx.New[PathValueGetter]()

func PathValueGetterFromContext(ctx context.Context) PathValueGetter {
	if g, ok := paramGetterContext.MayFrom(ctx); ok {
		return g
	}
	return Params{}
}

func ContextWithPathValueGetter(ctx context.Context, p PathValueGetter) context.Context {
	return paramGetterContext.Inject(ctx, p)
}

type Params map[string]string

func (d Params) PathValue(k string) string {
	v, _ := d[k]
	return v
}
