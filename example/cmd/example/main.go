package main

import (
	"context"

	"github.com/go-courier/logr"
	"github.com/go-courier/logr/slog"
	"github.com/octohelm/courier/example/apis"
	"github.com/octohelm/courier/pkg/courierhttp/handler/httprouter"
	"github.com/octohelm/courier/pkg/httputil"
)

func main() {
	ctx := logr.WithLogger(context.Background(), slog.Logger(slog.Default()))

	h, err := httprouter.New(apis.R, "example")
	if err != nil {
		panic(err)
	}

	if err := httputil.ListenAndServe(ctx, "0.0.0.0:9001", h); err != nil {
		panic(err)
	}
}
