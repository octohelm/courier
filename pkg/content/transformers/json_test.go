package transformers_test

import (
	"context"
	"reflect"
	"testing"

	. "github.com/octohelm/x/testing/v2"

	"github.com/octohelm/courier/internal/testingutil"
	"github.com/octohelm/courier/pkg/content/internal"
)

func TestJSONTransformerRoundTrip(t *testing.T) {
	type Data struct {
		A      string   `json:"a"`
		B      int      `json:"b"`
		Filter []string `json:"filter"`
	}

	op := struct {
		Body Data `in:"body"`
	}{
		Body: Data{
			A:      "str",
			B:      2,
			Filter: []string{"x1", "x2"},
		},
	}

	Then(t, "JSON body 可以在请求构造和反序列化之间保持一致",
		ExpectMust(func() error {
			req, err := internal.NewRequest(context.Background(), "POST", "/", op)
			if err != nil {
				return err
			}
			if err := testingutil.BeRequest(`
POST / HTTP/1.1
Content-Type: application/json; charset=utf-8

{"a":"str","b":2,"filter":["x1","x2"]}
`)(req); err != nil {
				return err
			}

			op2 := struct {
				Body Data `in:"body"`
			}{}

			if err := internal.UnmarshalRequest(req, &op2); err != nil {
				return err
			}
			if !reflect.DeepEqual(op2.Body, op.Body) {
				return errContent("unexpected decoded json body")
			}
			return nil
		}),
	)
}

func errContent(msg string) error {
	return &contentErr{msg: msg}
}

type contentErr struct{ msg string }

func (e *contentErr) Error() string { return e.msg }
