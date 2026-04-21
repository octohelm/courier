package httputil

import (
	"context"
	"net/http"
	"os"
	"syscall"
	"testing"
	"time"

	. "github.com/octohelm/x/testing/v2"
)

func TestListenAndServe(t0 *testing.T) {
	Then(t0, "服务可启动并在收到终止信号后优雅退出",
		ExpectMust(func() error {
			done := make(chan error, 1)

			go func() {
				done <- ListenAndServe(context.Background(), "127.0.0.1:0", http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
					rw.WriteHeader(http.StatusNoContent)
				}))
			}()

			time.Sleep(100 * time.Millisecond)

			if err := syscall.Kill(os.Getpid(), syscall.SIGTERM); err != nil {
				return err
			}

			select {
			case err := <-done:
				return err
			case <-time.After(3 * time.Second):
				return context.DeadlineExceeded
			}
		}),
	)
}
