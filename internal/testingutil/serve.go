package testingutil

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

func Serve(t testing.TB, handler http.Handler) *httptest.Server {
	srv := httptest.NewUnstartedServer(handler)
	srv.Start()
	t.Cleanup(srv.Close)
	return srv
}

func ServeWithH2C(t testing.TB, handler http.Handler) *httptest.Server {
	srv := httptest.NewUnstartedServer(h2c.NewHandler(handler, &http2.Server{}))
	srv.Start()
	t.Cleanup(srv.Close)
	return srv
}
