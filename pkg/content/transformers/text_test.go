package transformers_test

import (
	"context"
	"testing"

	. "github.com/octohelm/x/testing/v2"

	"github.com/octohelm/courier/internal/testingutil"
	"github.com/octohelm/courier/pkg/content/internal"
)

func TestTextTransformerRoundTrip(t *testing.T) {
	op := struct {
		Body string `in:"body"`
	}{
		Body: "test",
	}

	Then(t, "text body 可以在请求构造和反序列化之间保持一致", ExpectMust(func() error {
		req, err := internal.NewRequest(context.Background(), "POST", "/", op)
		if err != nil {
			return err
		}
		if err := testingutil.BeRequest(`
POST / HTTP/1.1
Content-Length: 4
Content-Type: text/plain; charset=utf-8

test
`)(req); err != nil {
			return err
		}

		op2 := struct {
			Body []byte `in:"body" mime:"text"`
		}{}

		if err := internal.UnmarshalRequest(req, &op2); err != nil {
			return err
		}
		if string(op2.Body) != "test" {
			return errContent("unexpected decoded text body")
		}
		return nil
	}))
}
