package plain

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"reflect"
	"testing"

	"github.com/octohelm/courier/pkg/transformer/core"

	"github.com/octohelm/x/ptr"
	typesutil "github.com/octohelm/x/types"
	. "github.com/onsi/gomega"
)

func TestTextTransformer(t *testing.T) {
	ct, _ := core.NewTransformer(context.Background(), typesutil.FromRType(reflect.TypeOf("")), core.Option{})

	t.Run("EncodeTo", func(t *testing.T) {
		t.Run("raw value", func(t *testing.T) {
			b := bytes.NewBuffer(nil)
			h := http.Header{}
			err := ct.EncodeTo(context.Background(), core.WriterWithHeader(b, h), "")
			NewWithT(t).Expect(err).To(BeNil())
			NewWithT(t).Expect(h.Get("Content-Type")).To(Equal("text/plain; charset=utf-8"))
		})

		t.Run("reflect value", func(t *testing.T) {
			b := bytes.NewBuffer(nil)
			h := http.Header{}
			err := ct.EncodeTo(context.Background(), core.WriterWithHeader(b, h), reflect.ValueOf(1))
			NewWithT(t).Expect(err).To(BeNil())
			NewWithT(t).Expect(h.Get("Content-Type")).To(Equal("text/plain; charset=utf-8"))
		})
	})

	t.Run("DecodeAndValidate", func(t *testing.T) {
		t.Run("failed", func(t *testing.T) {
			b := bytes.NewBufferString("a")
			i := 0
			err := ct.DecodeFrom(context.Background(), io.NopCloser(b), &i)
			NewWithT(t).Expect(err).NotTo(BeNil())
		})

		t.Run("success", func(t *testing.T) {
			b := bytes.NewBufferString("1")
			err := ct.DecodeFrom(context.Background(), io.NopCloser(b), reflect.ValueOf(ptr.Int(0)))
			NewWithT(t).Expect(err).To(BeNil())
		})
	})
}
