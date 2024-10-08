package transformers_test

import (
	"context"
	"testing"

	"github.com/octohelm/courier/internal/testingutil"
	"github.com/octohelm/courier/pkg/content/internal"
	testingx "github.com/octohelm/x/testing"
)

func TestOctetTransformer(t *testing.T) {
	op := struct {
		Body string `in:"body" mime:"octet"`
	}{
		Body: "test",
	}

	req, err := internal.NewRequest(context.Background(), "POST", "/", op)
	testingx.Expect(t, err, testingx.BeNil[error]())
	testingx.Expect(t, req, testingutil.BeRequest(`
POST / HTTP/1.1
Content-Type: application/octet-stream

test
`))

	op2 := struct {
		Body []byte `in:"body" mime:"octet"`
	}{}

	err = internal.UnmarshalRequest(req, &op2)
	testingx.Expect(t, err, testingx.BeNil[error]())
	testingx.Expect(t, string(op2.Body), testingx.Be("test"))
}
