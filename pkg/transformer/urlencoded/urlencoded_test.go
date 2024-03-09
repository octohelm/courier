package urlencoded_test

import (
	"bytes"
	"context"
	"encoding"
	"fmt"
	encodingx "github.com/octohelm/x/encoding"
	"github.com/pkg/errors"
	"io"
	"net/http"
	"reflect"
	"testing"

	"github.com/octohelm/courier/internal/testingutil"
	"github.com/octohelm/courier/pkg/transformer"
	"github.com/octohelm/courier/pkg/transformer/core"

	"github.com/octohelm/x/ptr"
	typesx "github.com/octohelm/x/types"
	. "github.com/onsi/gomega"
)

func TestTransformerURLEncoded(t *testing.T) {
	queryStr := `Bool=true` +
		`&Bytes=Ynl0ZXM%3D` +
		`&PtrInt=1` +
		`&StringArray=1&StringArray=&StringArray=3` +
		`&StringSlice=1&StringSlice=2&StringSlice=3` +
		`&StringSliceCommaJoin=1%2C2%2C3` +
		`&Struct=%3CSub%3E%3CName%3E%3C%2FName%3E%3C%2FSub%3E` +
		`&StructSlice=%7B%22Name%22%3A%22name%22%7D%0A` +
		`&first_name=test`

	type Sub struct {
		Name string
	}

	type TestData struct {
		PtrBool              *bool `name:",omitempty"`
		PtrInt               *int
		Bool                 bool
		Bytes                []byte
		FirstName            string `name:"first_name"`
		StructSlice          []Sub
		StringSlice          []string
		StringSliceCommaJoin CommaJoin[string]
		StringArray          [3]string
		Struct               Sub `mime:"xml"`
	}

	data := TestData{}
	data.FirstName = "test"
	data.Bool = true
	data.Bytes = []byte("bytes")
	data.PtrInt = ptr.Int(1)
	data.StringSlice = []string{"1", "2", "3"}
	data.StringSliceCommaJoin = []string{"1", "2", "3"}
	data.StructSlice = []Sub{
		{
			Name: "name",
		},
	}
	data.StringArray = [3]string{"1", "", "3"}

	ct, err := transformer.NewTransformer(context.Background(), typesx.FromRType(reflect.TypeOf(data)), core.Option{
		MIME: "urlencoded",
	})
	testingutil.Expect(t, err, testingutil.Be[error](nil))

	t.Run("EncodeTo", func(t *testing.T) {
		b := bytes.NewBuffer(nil)
		h := http.Header{}

		err := ct.EncodeTo(context.Background(), core.WriterWithHeader(b, h), data)

		NewWithT(t).Expect(err).To(BeNil())
		NewWithT(t).Expect(h.Get("Content-Type")).To(Equal("application/x-www-form-urlencoded; param=value"))
		NewWithT(t).Expect(b.String()).To(Equal(queryStr))
	})

	t.Run("DecodeAndValidate", func(t *testing.T) {
		b := bytes.NewBufferString(queryStr)
		testData := TestData{}

		err := ct.DecodeFrom(context.Background(), io.NopCloser(b), &testData)
		NewWithT(t).Expect(err).To(BeNil())
		NewWithT(t).Expect(testData).To(Equal(data))
	})
}

type CommaJoin[T any] []T

var _ encoding.TextUnmarshaler = &CommaJoin[string]{}
var _ encoding.TextMarshaler = &CommaJoin[string]{}

func (x *CommaJoin[T]) UnmarshalText(data []byte) error {
	parts := bytes.Split(data, []byte(","))
	if n := len(parts); n > 0 {
		list := make(CommaJoin[T], n)

		for i := range list {
			x := list[i]
			if err := encodingx.UnmarshalText(&x, parts[i]); err != nil {
				return errors.Wrapf(err, "unmarshal failed: %s", string(parts[i]))
			}
			list[i] = x
		}

		*x = list
	}
	return nil
}

func (list CommaJoin[T]) MarshalText() ([]byte, error) {
	b := bytes.NewBuffer(nil)
	for i, x := range list {
		if i > 0 {
			b.WriteString(",")
		}
		_, _ = fmt.Fprintf(b, "%v", x)
	}
	return b.Bytes(), nil
}
