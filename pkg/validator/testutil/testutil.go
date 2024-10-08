package testutil

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/go-json-experiment/json/jsontext"

	"github.com/go-json-experiment/json"
	"github.com/octohelm/courier/pkg/validator/internal"
	testingx "github.com/octohelm/x/testing"
)

type Cases = []Case

type Case struct {
	Input       []byte
	Target      any
	Expect      []byte
	ExpectError func(err error, v any) bool
}

func Run(t *testing.T, cases ...Case) {
	for i, tc := range cases {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			input := tc.Input
			if input == nil {
				input = tc.Expect
			}
			err := internal.UnmarshalDecode(jsontext.NewDecoder(bytes.NewBuffer(input)), tc.Target)
			if tc.ExpectError != nil {
				testingx.Expect(t, tc.ExpectError(err, tc.Target), testingx.BeTrue())
			} else {
				testingx.Expect(t, err, testingx.BeNil[error]())
			}

			if tc.Expect != nil {
				out, err := json.Marshal(tc.Target)
				testingx.Expect(t, err, testingx.BeNil[error]())
				testingx.Expect(t, string(bytes.TrimSpace(out)), testingx.Be(string(bytes.TrimSpace(tc.Expect))))
			}
		})
	}
}
