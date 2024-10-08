package testingutil

import (
	"bytes"
	"context"
	"fmt"
	"github.com/octohelm/courier/pkg/courierhttp/transport"
	testingx "github.com/octohelm/x/testing"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
)

func ShouldReturnWhenRequest(req any, expect string) testingx.Matcher[http.Handler] {
	r, err := transport.NewRequest(context.Background(), req)
	if err != nil {
		panic(err)
	}

	return &responseMatcher{
		req:    r,
		expect: unifyRequestData([]byte(expect)),
	}
}

type responseMatcher struct {
	req      *http.Request
	expect   []byte
	reqData  []byte
	respData []byte
}

func (r *responseMatcher) Name() string {
	return "Return When Request"
}

func (r *responseMatcher) Negative() bool {
	return false
}

func (r *responseMatcher) Match(h http.Handler) bool {
	rw := NewMockResponseWriter()
	h.ServeHTTP(rw, r.req)
	r.respData = unifyRequestData(rw.MustDumpResponse())

	return bytes.Equal(r.respData, r.expect)
}

func (r *responseMatcher) FormatActual(actual http.Handler) string {
	fmt.Println(string(r.respData))

	return string(r.respData)
}

func (m *responseMatcher) FormatExpected() string {
	return string(m.expect)
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
	return data
}
