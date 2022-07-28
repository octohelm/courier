package testingutil

import (
	"bytes"
	"context"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"strings"
	"testing"

	"github.com/octohelm/courier/pkg/courierhttp/transport"
)

func ShouldReturnWhenRequest(t testing.TB, h http.Handler, req any, httpResp string) {
	t.Helper()

	r, err := transport.NewRequest(context.Background(), req)
	Expect(t, err, Be[error](nil))

	rw := NewMockResponseWriter()
	h.ServeHTTP(rw, r)
	Expect(t, strings.TrimSpace(string(rw.MustDumpResponse())), Be(strings.TrimSpace(httpResp)))
}

func NewMockResponseWriter() *MockResponseWriter {
	return &MockResponseWriter{
		header: http.Header{},
	}
}

type MockResponseWriter struct {
	header     http.Header
	StatusCode int
	bytes.Buffer
}

var _ http.ResponseWriter = (*MockResponseWriter)(nil)

func (w *MockResponseWriter) Header() http.Header {
	if w.StatusCode == 0 {
		return w.header
	}

	header := http.Header{}

	for k, v := range w.header {
		header[k] = v
	}

	return header
}

func (w *MockResponseWriter) WriteHeader(statusCode int) {
	w.StatusCode = statusCode
}

func (w *MockResponseWriter) Response() *http.Response {
	resp := &http.Response{}
	resp.Header = w.header
	resp.StatusCode = w.StatusCode
	resp.Body = ioutil.NopCloser(&w.Buffer)
	return resp
}

func (w *MockResponseWriter) MustDumpResponse() []byte {
	data, err := httputil.DumpResponse(w.Response(), true)
	if err != nil {
		panic(err)
	}
	return bytes.Replace(data, []byte("\r\n"), []byte("\n"), -1)
}
