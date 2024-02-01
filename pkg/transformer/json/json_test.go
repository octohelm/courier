package json

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"reflect"
	"testing"

	"github.com/octohelm/courier/pkg/transformer/core"

	verrors "github.com/octohelm/courier/pkg/validator"
	typesutil "github.com/octohelm/x/types"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
)

type S string

func (s *S) UnmarshalText(data []byte) error {
	return errors.Errorf("err")
}

func TestJSONTransformer(t *testing.T) {
	data := struct {
		Data struct {
			S           S    `json:"s,omitempty"`
			Bool        bool `json:"bool"`
			StructSlice []struct {
				Name string `json:"name"`
			} `json:"structSlice"`
			StringSlice []string `json:"stringSlice"`
			NestedSlice []struct {
				Names []string `json:"names"`
			} `json:"nestedSlice"`
		} `json:"data"`
	}{}

	ct, _ := core.NewTransformer(context.Background(), typesutil.FromRType(reflect.TypeOf(data)), core.Option{})

	t.Run("EncodeTo", func(t *testing.T) {
		b := bytes.NewBuffer(nil)
		h := http.Header{}

		err := ct.EncodeTo(context.Background(), core.WriterWithHeader(b, h), data)
		NewWithT(t).Expect(err).To(BeNil())
		NewWithT(t).Expect(h.Get("Content-Type")).To(Equal("application/json; charset=utf-8"))
	})

	t.Run("EncodeTo with reflect.value", func(t *testing.T) {
		b := bytes.NewBuffer(nil)
		h := http.Header{}

		err := ct.EncodeTo(context.Background(), core.WriterWithHeader(b, h), reflect.ValueOf(data))
		NewWithT(t).Expect(err).To(BeNil())
		NewWithT(t).Expect(h.Get("Content-Type")).To(Equal("application/json; charset=utf-8"))
	})

	t.Run("DecodeAndValidate failed", func(t *testing.T) {
		b := io.NopCloser(bytes.NewBufferString(`{`))
		err := ct.DecodeFrom(context.Background(), b, &data)
		NewWithT(t).Expect(err).NotTo(BeNil())
	})

	t.Run("DecodeAndValidate success", func(t *testing.T) {
		b := io.NopCloser(bytes.NewBufferString(`{}`))
		err := ct.DecodeFrom(context.Background(), b, reflect.ValueOf(&data))
		NewWithT(t).Expect(err).To(BeNil())
	})

	t.Run("DecodeAndValidate failed with location", func(t *testing.T) {
		cases := []struct {
			json     string
			location string
		}{{
			`{
	"data": {
		"s": "111",
		"bool": true
	}
}`, "data.s",
		},
			{
				`
{
 	"data": {
		"bool": ""
	}
}
`, "data.bool",
			},
			{
				`
{
		"data": {
			"structSlice": [
				{"name":"{"},
				{"name":"1"},
				{"name": { "test": 1 }},
				{"name":"1"}
			]
		}
}`,
				"data.structSlice[2].name",
			},
			{
				`
		{
			"data": {
				"stringSlice":["1","2",3]
			}
		}`,
				"data.stringSlice[2]",
			},
			{
				`
		{
			"data": {
				"stringSlice":["1","2",3]
			}
		}`,
				"data.stringSlice[2]",
			},
			{
				`
		{
			"data": {
				"bool": true,
				"nestedSlice": [
					{ "names": ["1","2","3"] },
			        { "names": ["1","\"2", 3] }
				]
			}
		}
		`, "data.nestedSlice[1].names[2]",
			},
		}

		for _, c := range cases {
			b := bytes.NewBufferString(c.json)
			err := ct.DecodeFrom(context.Background(), io.NopCloser(b), &data)

			err.(*verrors.ErrorSet).Each(func(fieldErr *verrors.FieldError) {
				NewWithT(t).Expect(fieldErr.Path.String()).To(Equal(c.location))
			})
		}
	})
}
