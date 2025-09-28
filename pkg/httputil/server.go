package httputil

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/octohelm/x/logr"
)

func ListenAndServe(ctx context.Context, addr string, handler http.Handler) error {
	logger := logr.FromContext(ctx)

	srv := &http.Server{
		ReadHeaderTimeout: 30 * time.Second,
		Addr:              addr,
		Handler:           handler,
	}

	go func() {
		logger.Info("listen on %s", addr)

		if err := srv.ListenAndServe(); err != nil {
			logger.Error(err)

			if !errors.Is(err, http.ErrServerClosed) {
				panic(err)
			}
		}
	}()

	stopCh := make(chan os.Signal, 1)
	signal.Notify(stopCh, os.Interrupt, syscall.SIGTERM)
	<-stopCh

	timeout := 10 * time.Second

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	logger.Info("shutdowning in %s", timeout)

	return srv.Shutdown(ctx)
}
