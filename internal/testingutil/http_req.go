package testingutil

import (
	"bytes"
	"net/http"
	"net/http/httputil"
	"regexp"

	testingx "github.com/octohelm/x/testing"
)

func BeRequest(expect string) testingx.Matcher[*http.Request] {
	return &requestMatcher{
		expect: unifyRequestData([]byte(expect)),
	}
}

type requestMatcher struct {
	expect []byte
	actual []byte
}

func (m *requestMatcher) Match(req *http.Request) bool {
	raw, _ := httputil.DumpRequest(req, true)
	m.actual = unifyRequestData(raw)
	return bytes.Equal(m.actual, m.expect)
}

func (m *requestMatcher) Action() string {
	return "Be Request"
}

func (m *requestMatcher) Negative() bool {
	return false
}

func (m *requestMatcher) NormalizeActual(req *http.Request) any {
	return string(m.actual)
}

var _ testingx.MatcherWithNormalizedExpected = &requestMatcher{}

func (m *requestMatcher) NormalizedExpected() any {
	return string(m.expect)
}

var reContentTypeWithBoundary = regexp.MustCompile(`Content-Type: multipart/form-data; boundary=([A-Za-z0-9]+)`)

func unifyRequestData(data []byte) []byte {
	data = bytes.Replace(data, []byte("\r\n"), []byte("\n"), -1)

	if reContentTypeWithBoundary.Match(data) {
		matches := reContentTypeWithBoundary.FindAllSubmatch(data, 1)
		data = bytes.Replace(data, matches[0][1], []byte("boundary1"), -1)
	}

	return bytes.TrimSpace(data)
}
