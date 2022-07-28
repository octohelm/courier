package httputil

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-courier/logr"
)

func ListenAndServe(ctx context.Context, addr string, handler http.Handler) error {
	logger := logr.FromContext(ctx)

	srv := &http.Server{Addr: addr, Handler: handler}

	go func() {
		logger.Info("listen on %s", addr)

		if err := srv.ListenAndServe(); err != nil {
			if err == http.ErrServerClosed {
				logger.Error(err)
			} else {
				logger.Fatal(err)
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
