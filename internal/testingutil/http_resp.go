package testingutil

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"maps"
	"net/http"
	"net/http/httputil"

	"github.com/octohelm/courier/pkg/courierhttp/transport"
)

func ShouldReturnWhenRequest(req any, expect string) func(http.Handler) error {
	r, err := transport.NewRequest(context.Background(), req)
	if err != nil {
		panic(err)
	}

	m := &responseMatcher{
		req:    r,
		expect: unifyRequestData([]byte(expect)),
	}

	return func(h http.Handler) error {
		rw := NewMockResponseWriter()
		h.ServeHTTP(rw, m.req)
		m.respData = unifyRequestData(rw.MustDumpResponse())

		if bytes.Equal(m.respData, m.expect) {
			return nil
		}

		return fmt.Errorf("response mismatch\nexpect:\n%s\nactual:\n%s", m.expect, m.respData)
	}
}

type responseMatcher struct {
	req      *http.Request
	expect   []byte
	reqData  []byte
	respData []byte
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

	maps.Copy(header, w.header)

	return header
}

func (w *MockResponseWriter) WriteHeader(statusCode int) {
	w.StatusCode = statusCode
}

func (w *MockResponseWriter) Response() *http.Response {
	resp := &http.Response{}
	resp.Header = w.header
	resp.StatusCode = w.StatusCode
	resp.Body = io.NopCloser(&w.Buffer)
	return resp
}

func (w *MockResponseWriter) MustDumpResponse() []byte {
	data, err := httputil.DumpResponse(w.Response(), true)
	if err != nil {
		panic(err)
	}
	return data
}
