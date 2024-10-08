package testingutil

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

func Serve(t testing.TB, handler http.Handler) *httptest.Server {
	srv := httptest.NewServer(h2c.NewHandler(handler, &http2.Server{}))
	t.Cleanup(func() {
		srv.Close()
	})
	return srv
}
