package transformers_test

import (
	"context"
	"reflect"
	"testing"

	. "github.com/octohelm/x/testing/v2"

	"github.com/octohelm/courier/internal/testingutil"
	"github.com/octohelm/courier/pkg/content/internal"
)

func TestURLEncodedTransformerRoundTrip(t *testing.T) {
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

	Then(t, "urlencoded body 可以在请求构造和反序列化之间保持一致", ExpectMust(func() error {
		req, err := internal.NewRequest(context.Background(), "POST", "/", op)
		if err != nil {
			return err
		}
		if err := testingutil.BeRequest(`
POST / HTTP/1.1
Content-Type: application/x-www-form-urlencoded; param=value

a=s&b=2&filter=x1&filter=x2
`)(req); err != nil {
			return err
		}

		op2 := struct {
			Body Data `in:"body" mime:"urlencoded"`
		}{}

		if err := internal.UnmarshalRequest(req, &op2); err != nil {
			return err
		}
		if !reflect.DeepEqual(op2.Body, op.Body) {
			return errContent("unexpected decoded urlencoded body")
		}
		return nil
	}))
}
