package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/octohelm/x/testing/v2"
)

func TestHandlerHelpers(t *testing.T) {
	Then(t, "中间件组合与路径参数辅助方法符合预期",
		ExpectMust(func() error {
			order := make([]string, 0, 3)
			final := ApplyMiddlewares(
				func(next http.Handler) http.Handler {
					return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
						order = append(order, "mw1-before")
						next.ServeHTTP(rw, req)
						order = append(order, "mw1-after")
					})
				},
				func(next http.Handler) http.Handler {
					return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
						order = append(order, "mw2-before")
						next.ServeHTTP(rw, req)
						order = append(order, "mw2-after")
					})
				},
			)(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
				order = append(order, "final")
				rw.WriteHeader(http.StatusNoContent)
			}))

			rec := httptest.NewRecorder()
			final.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

			if rec.Code != http.StatusNoContent {
				return errHandler("unexpected status code")
			}
			if len(order) != 5 {
				return errHandler("unexpected middleware order size")
			}
			expected := []string{"mw1-before", "mw2-before", "final", "mw2-after", "mw1-after"}
			for i := range expected {
				if order[i] != expected[i] {
					return errHandler("unexpected middleware order")
				}
			}
			return nil
		}),
		ExpectMust(func() error {
			ctx := ContextWithPathValueGetter(context.Background(), Params{
				"id":   "1",
				"name": "demo",
			})

			getter := PathValueGetterFromContext(ctx)
			if getter == nil {
				return errHandler("missing getter")
			}
			if getter.PathValue("id") != "1" {
				return errHandler("unexpected path value")
			}
			if getter.PathValue("name") != "demo" {
				return errHandler("unexpected helper path value")
			}
			if getter.PathValue("missing") != "" {
				return errHandler("unexpected missing path value")
			}
			return nil
		}),
	)
}

func errHandler(msg string) error {
	return &handlerErr{msg: msg}
}

type handlerErr struct{ msg string }

func (e *handlerErr) Error() string { return e.msg }
