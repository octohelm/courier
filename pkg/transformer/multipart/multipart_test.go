package multipart_test

import (
	"bytes"
	"context"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"reflect"
	"strings"
	"testing"

	"github.com/octohelm/courier/pkg/courierhttp"

	"github.com/octohelm/courier/pkg/transformer"
	"github.com/octohelm/courier/pkg/transformer/core"
	multiparttransformer "github.com/octohelm/courier/pkg/transformer/multipart"

	"github.com/octohelm/x/ptr"
	typesutil "github.com/octohelm/x/types"
	. "github.com/onsi/gomega"
)

func TestMultipartTransformer(t *testing.T) {
	parts := `--99bb5d156e61cf661d01fc370479b62a3451759d25d14711fd7e9db170f6
Content-Disposition: form-data; name="PtrBool"
Content-Type: text/plain; charset=utf-8

true
--99bb5d156e61cf661d01fc370479b62a3451759d25d14711fd7e9db170f6
Content-Disposition: form-data; name="PtrInt"
Content-Type: text/plain; charset=utf-8

1
--99bb5d156e61cf661d01fc370479b62a3451759d25d14711fd7e9db170f6
Content-Disposition: form-data; name="Bool"
Content-Type: text/plain; charset=utf-8

true
--99bb5d156e61cf661d01fc370479b62a3451759d25d14711fd7e9db170f6
Content-Disposition: form-data; name="bytes"
Content-Type: text/plain; charset=utf-8

Ynl0ZXM=
--99bb5d156e61cf661d01fc370479b62a3451759d25d14711fd7e9db170f6
Content-Disposition: form-data; name="first_name"
Content-Type: text/plain; charset=utf-8

test
--99bb5d156e61cf661d01fc370479b62a3451759d25d14711fd7e9db170f6
Content-Disposition: form-data; name="StructSlice"
Content-Type: application/json; charset=utf-8

{"Name":"name"}

--99bb5d156e61cf661d01fc370479b62a3451759d25d14711fd7e9db170f6
Content-Disposition: form-data; name="StringSlice"
Content-Type: text/plain; charset=utf-8

1
--99bb5d156e61cf661d01fc370479b62a3451759d25d14711fd7e9db170f6
Content-Disposition: form-data; name="StringSlice"
Content-Type: text/plain; charset=utf-8

2
--99bb5d156e61cf661d01fc370479b62a3451759d25d14711fd7e9db170f6
Content-Disposition: form-data; name="StringSlice"
Content-Type: text/plain; charset=utf-8

3
--99bb5d156e61cf661d01fc370479b62a3451759d25d14711fd7e9db170f6
Content-Disposition: form-data; name="StringArray"
Content-Type: text/plain; charset=utf-8

1
--99bb5d156e61cf661d01fc370479b62a3451759d25d14711fd7e9db170f6
Content-Disposition: form-data; name="StringArray"
Content-Type: text/plain; charset=utf-8


--99bb5d156e61cf661d01fc370479b62a3451759d25d14711fd7e9db170f6
Content-Disposition: form-data; name="StringArray"
Content-Type: text/plain; charset=utf-8

3
--99bb5d156e61cf661d01fc370479b62a3451759d25d14711fd7e9db170f6
Content-Disposition: form-data; name="Struct"
Content-Type: application/xml; charset=utf-8

<Sub><Name></Name></Sub>
--99bb5d156e61cf661d01fc370479b62a3451759d25d14711fd7e9db170f6
Content-Disposition: form-data; name="Files"; filename="file0.txt"
Content-Type: application/octet-stream

text0
--99bb5d156e61cf661d01fc370479b62a3451759d25d14711fd7e9db170f6
Content-Disposition: form-data; name="Files"; filename="file1.txt"
Content-Type: application/octet-stream

text1
--99bb5d156e61cf661d01fc370479b62a3451759d25d14711fd7e9db170f6
Content-Disposition: form-data; name="File"; filename="file.txt"
Content-Type: application/octet-stream

text
--99bb5d156e61cf661d01fc370479b62a3451759d25d14711fd7e9db170f6--`

	type Sub struct {
		Name string
	}

	type TestData struct {
		PtrBoolEmpty *bool `name:",omitempty"`
		PtrBool      *bool `name:",omitempty"`
		PtrInt       *int
		Bool         bool
		Bytes        []byte `name:"bytes"`
		FirstName    string `name:"first_name,omitempty"`
		StructSlice  []Sub
		StringSlice  []string
		StringArray  [3]string
		Struct       Sub             `mime:"xml"`
		Files        []io.ReadCloser `name:",omitempty"`
		File         io.ReadCloser   `name:",omitempty"`
	}

	data := TestData{}
	data.PtrBool = ptr.Bool(true)
	data.FirstName = "test"
	data.Bool = true
	data.Bytes = []byte("bytes")
	data.PtrInt = ptr.Int(1)
	data.StringSlice = []string{"1", "2", "3"}
	data.StructSlice = []Sub{
		{
			Name: "name",
		},
	}
	data.StringArray = [3]string{"1", "", "3"}

	data.File = multiparttransformer.WrapFileHeader(
		io.NopCloser(bytes.NewBufferString("text")),
		multiparttransformer.WithFilename("file.txt"),
		multiparttransformer.WithName("File"),
	)

	data.Files = []io.ReadCloser{
		multiparttransformer.WrapFileHeader(
			io.NopCloser(bytes.NewBufferString("text0")),
			multiparttransformer.WithFilename("file0.txt"),
			multiparttransformer.WithName("Files"),
		),
		multiparttransformer.WrapFileHeader(
			io.NopCloser(bytes.NewBufferString("text1")),
			multiparttransformer.WithFilename("file1.txt"),
			multiparttransformer.WithName("Files"),
		),
	}

	ct, _ := transformer.NewTransformer(context.Background(), typesutil.FromRType(reflect.TypeOf(data)), core.Option{
		MIME: "multipart",
	})

	t.Run("EncodeTo", func(t *testing.T) {
		b := bytes.NewBuffer(nil)
		h := http.Header{}
		err := ct.EncodeTo(context.Background(), core.WriterWithHeader(b, h), data)
		NewWithT(t).Expect(err).To(BeNil())
		_, params, _ := mime.ParseMediaType(h.Get("Content-Type"))

		gen := toParts(b, params["boundary"])
		expect := toParts(bytes.NewBufferString(replaceBoundaryMultipart(parts, params["boundary"])), params["boundary"])

		NewWithT(t).Expect(len(gen)).To(Equal(len(expect)))

		for i := range gen {
			NewWithT(t).Expect(gen[i].FormName()).To(Equal(expect[i].FormName()))
			NewWithT(t).Expect(gen[i].FileName()).To(Equal(expect[i].FileName()))
			NewWithT(t).Expect(gen[i].Header).To(Equal(expect[i].Header))
		}
	})

	t.Run("DecodeAndValidate", func(t *testing.T) {
		b := io.NopCloser(bytes.NewBufferString(parts))
		testData := TestData{}

		err := ct.DecodeFrom(context.Background(), b, &testData, textproto.MIMEHeader{
			"Content-type": []string{
				mime.FormatMediaType(ct.Names()[0], map[string]string{
					"boundary": boundary,
				}),
			},
		})

		NewWithT(t).Expect(err).To(BeNil())

		NewWithT(t).Expect(testData.PtrBoolEmpty).To(Equal(data.PtrBoolEmpty))
		NewWithT(t).Expect(testData.PtrBool).To(Equal(data.PtrBool))
		NewWithT(t).Expect(testData.PtrInt).To(Equal(data.PtrInt))
		NewWithT(t).Expect(testData.Bool).To(Equal(data.Bool))
		NewWithT(t).Expect(testData.Bytes).To(Equal(data.Bytes))
		NewWithT(t).Expect(testData.FirstName).To(Equal(data.FirstName))
		NewWithT(t).Expect(testData.StructSlice).To(Equal(data.StructSlice))
		NewWithT(t).Expect(testData.StringSlice).To(Equal(data.StringSlice))
		NewWithT(t).Expect(testData.StringArray).To(Equal(data.StringArray))
		NewWithT(t).Expect(testData.Struct).To(Equal(data.Struct))

		for i := range testData.Files {
			NewWithT(t).Expect(
				testData.Files[i].(courierhttp.FileHeader).Filename(),
			).To(Equal(data.Files[i].(courierhttp.FileHeader).Filename()))
		}
	})
}

var boundary = "99bb5d156e61cf661d01fc370479b62a3451759d25d14711fd7e9db170f6"

func replaceBoundaryMultipart(data string, generatedBoundary string) string {
	return strings.Replace(data, boundary, generatedBoundary, -1)
}

func toParts(r io.Reader, b string) (parts []*multipart.Part) {
	gen := multipart.NewReader(r, b)

	for {
		rp, err := gen.NextPart()
		if err != nil {
			break
		}
		data := bytes.NewBuffer(nil)
		_, _ = io.Copy(data, rp)
		rp.Header["Content"] = []string{data.String()}
		_ = rp.Close()
		parts = append(parts, rp)
	}

	return
}
