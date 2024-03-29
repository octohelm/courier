package testingutil

import (
	"bytes"
	"net/http/httptest"
	"regexp"
	"testing"

	"net/http"
	"net/http/httputil"

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

func RequestEqual(t testing.TB, req *http.Request, expect string) {
	t.Helper()

	data, err := httputil.DumpRequest(req, true)
	Expect(t, err, Be[error](nil))
	Expect(t, string(unifyRequestData(data)), Equal(string(unifyRequestData([]byte(expect)))))
}

var reContentTypeWithBoundary = regexp.MustCompile(`Content-Type: multipart/form-data; boundary=([A-Za-z0-9]+)`)

func unifyRequestData(data []byte) []byte {
	data = bytes.Replace(data, []byte("\r\n"), []byte("\n"), -1)

	if reContentTypeWithBoundary.Match(data) {
		matches := reContentTypeWithBoundary.FindAllSubmatch(data, 1)
		data = bytes.Replace(data, matches[0][1], []byte("boundary1"), -1)
	}

	return data
}
