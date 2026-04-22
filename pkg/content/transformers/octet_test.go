package transformers_test

import (
	"context"
	"testing"

	. "github.com/octohelm/x/testing/v2"

	"github.com/octohelm/courier/internal/testingutil"
	"github.com/octohelm/courier/pkg/content/internal"
)

func TestOctetTransformerRoundTrip(t *testing.T) {
	op := struct {
		Body string `in:"body" mime:"octet"`
	}{
		Body: "test",
	}

	Then(t, "octet-stream body 可以在请求构造和反序列化之间保持一致", ExpectMust(func() error {
		req, err := internal.NewRequest(context.Background(), "POST", "/", op)
		if err != nil {
			return err
		}
		if err := testingutil.BeRequest(`
POST / HTTP/1.1
Content-Length: 4
Content-Type: application/octet-stream

test
`)(req); err != nil {
			return err
		}

		op2 := struct {
			Body []byte `in:"body" mime:"octet"`
		}{}

		if err := internal.UnmarshalRequest(req, &op2); err != nil {
			return err
		}
		if string(op2.Body) != "test" {
			return errContent("unexpected decoded octet body")
		}
		return nil
	}))
}
