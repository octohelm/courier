package handler

import "net/http"

func ApplyMiddlewares(mw ...Middleware) Middleware {
	return func(final http.Handler) http.Handler {
		last := final
		for i := len(mw) - 1; i >= 0; i-- {
			last = mw[i](last)
		}
		return last
	}
}

type Middleware = func(http.Handler) http.Handler
