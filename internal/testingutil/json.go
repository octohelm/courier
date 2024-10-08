package testingutil

import (
	"bytes"
	"fmt"
	"os"

	"github.com/go-json-experiment/json"
	testingx "github.com/octohelm/x/testing"
)

func PrintJSON(v interface{}) {
	if err := json.MarshalWrite(os.Stdout, v); err != nil {
		panic(err)
	}
}

func BeJSON[X any](expect string) testingx.Matcher[X] {
	return &jsonMatcher[X]{
		expect: bytes.TrimSpace([]byte(expect)),
	}
}

type jsonMatcher[X any] struct {
	expect []byte
	actual []byte
}

func (m *jsonMatcher[X]) Match(v X) bool {
	m.actual, _ = json.Marshal(v)

	return bytes.Equal(m.actual, m.expect)
}

func (m *jsonMatcher[X]) Negative() bool {
	return false
}

func (m *jsonMatcher[X]) FormatActual(x X) string {
	fmt.Println(string(m.actual))

	return string(m.actual)
}

func (m *jsonMatcher[X]) FormatExpected() string {
	return string(m.expect)
}

func (m *jsonMatcher[X]) Name() string {
	return "Be JSON"
}
