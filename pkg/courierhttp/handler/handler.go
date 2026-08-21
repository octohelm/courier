package handler

import (
	"net/http"
	"slices"
)

func ApplyMiddlewares(mw ...Middleware) Middleware {
	return func(final http.Handler) http.Handler {
		last := final
		for _, m := range slices.Backward(mw) {
			last = m(last)
		}
		return last
	}
}

type Middleware = func(http.Handler) http.Handler
