package handler

import (
	"context"

	"github.com/octohelm/courier/internal/httprequest"
)

type PathValueGetter = httprequest.PathValueGetter

func PathValueGetterFromContext(ctx context.Context) PathValueGetter {
	return httprequest.PathValueGetterFromContext(ctx)
}

func ContextWithPathValueGetter(ctx context.Context, p PathValueGetter) context.Context {
	return httprequest.ContextWithPathValueGetter(ctx, p)
}

type Params map[string]string

func (d Params) PathValue(k string) string {
	v, _ := d[k]
	return v
}
