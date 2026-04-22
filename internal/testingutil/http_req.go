package testingutil

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httputil"
	"regexp"
)

func BeRequest(expect string) func(req *http.Request) error {
	expectData := unifyRequestData([]byte(expect))

	return func(req *http.Request) error {
		raw, _ := httputil.DumpRequest(req, true)
		actual := unifyRequestData(raw)

		if bytes.Equal(actual, expectData) {
			return nil
		}

		return fmt.Errorf("request mismatch\nexpect:\n%s\nactual:\n%s", expectData, actual)
	}
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
