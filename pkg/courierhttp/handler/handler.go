package handler

import "net/http"

func ApplyHandlerMiddlewares(mw ...HandlerMiddleware) HandlerMiddleware {
	return func(final http.Handler) http.Handler {
		last := final
		for i := len(mw) - 1; i >= 0; i-- {
			last = mw[i](last)
		}
		return last
	}
}

type HandlerMiddleware = func(http.Handler) http.Handler
