package transformers_test

import (
	"context"
	"testing"

	"github.com/octohelm/courier/internal/testingutil"
	"github.com/octohelm/courier/pkg/content/internal"
	testingx "github.com/octohelm/x/testing"
)

func TestJsonTransformer(t *testing.T) {
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

	req, err := internal.NewRequest(context.Background(), "POST", "/", op)
	testingx.Expect(t, err, testingx.BeNil[error]())
	testingx.Expect(t, req, testingutil.BeRequest(`
POST / HTTP/1.1
Content-Type: application/json; charset=utf-8

{"a":"str","b":2,"filter":["x1","x2"]}
`))

	op2 := struct {
		Body Data `in:"body"`
	}{}

	err = internal.UnmarshalRequest(req, &op2)
	testingx.Expect(t, err, testingx.BeNil[error]())
	testingx.Expect(t, op2.Body, testingx.Equal(op.Body))
}
