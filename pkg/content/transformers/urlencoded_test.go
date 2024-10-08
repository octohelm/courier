package transformers_test

import (
	"context"
	"testing"

	"github.com/octohelm/courier/internal/testingutil"
	"github.com/octohelm/courier/pkg/content/internal"
	testingx "github.com/octohelm/x/testing"
)

func TestUrlEncodedTransformer(t *testing.T) {
	type Data struct {
		A      string   `json:"a"`
		B      int      `json:"b"`
		Filter []string `json:"filter"`
	}

	op := struct {
		Body Data `in:"body" mime:"urlencoded"`
	}{
		Body: Data{
			A:      "s",
			B:      2,
			Filter: []string{"x1", "x2"},
		},
	}

	req, err := internal.NewRequest(context.Background(), "POST", "/", op)
	testingx.Expect(t, err, testingx.BeNil[error]())
	testingx.Expect(t, req, testingutil.BeRequest(`
POST / HTTP/1.1
Content-Type: application/x-www-form-urlencoded; param=value

a=s&b=2&filter=x1&filter=x2
`))

	op2 := struct {
		Body Data `in:"body" mime:"urlencoded"`
	}{}

	err = internal.UnmarshalRequest(req, &op2)
	testingx.Expect(t, err, testingx.BeNil[error]())
	testingx.Expect(t, op2.Body, testingx.Equal(op.Body))
}
