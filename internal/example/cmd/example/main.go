package main

import (
	"context"
	"net/http"

	orgmem "github.com/octohelm/courier/internal/example/domain/org/service/mem"
	orgservice "github.com/octohelm/courier/internal/example/domain/org/service"
	storemem "github.com/octohelm/courier/internal/example/domain/store/service/mem"
	storeservice "github.com/octohelm/courier/internal/example/domain/store/service"
	exampleroutes "github.com/octohelm/courier/internal/example/cmd/example/routes"
	"github.com/octohelm/courier/pkg/courierhttp/handler"
	"github.com/octohelm/courier/pkg/courierhttp/handler/httprouter"
	"github.com/octohelm/courier/pkg/httputil"
	"github.com/octohelm/x/logr"
	"github.com/octohelm/x/logr/slog"
)

func main() {
	ctx := logr.WithLogger(context.Background(), slog.Logger(slog.Default()))

	h, err := httprouter.New(exampleroutes.R, "example")
	if err != nil {
		panic(err)
	}

	orgSvc := orgmem.New()
	storeSvc := storemem.New()

	h = handler.ApplyMiddlewares(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			ctx := orgservice.ServiceInjectContext(req.Context(), orgSvc)
			ctx = storeservice.ServiceInjectContext(ctx, storeSvc)
			next.ServeHTTP(rw, req.WithContext(ctx))
		})
	})(h)

	if err := httputil.ListenAndServe(ctx, "0.0.0.0:9001", h); err != nil {
		panic(err)
	}
}
