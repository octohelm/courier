package transformers_test

import (
	"context"
	"io"
	"testing"

	"github.com/octohelm/courier/internal/testingutil"
	"github.com/octohelm/courier/pkg/content/internal"
	testingx "github.com/octohelm/x/testing"
)

func TestMultipartTransformer(t *testing.T) {
	type Data struct {
		A      string   `json:"a"`
		Filter []string `json:"filter"`
		File   File     `json:"file"`
		Files  []File   `json:"files"`
	}

	op := struct {
		Body Data `in:"body" mime:"multipart"`
	}{
		Body: Data{
			A:      "s",
			Filter: []string{"x1", "x2"},
			File: File{
				Name: "1.txt",
				Type: "text/plain",
				Data: []byte("text"),
			},
			Files: []File{
				{
					Name: "2.txt",
					Type: "text/plain",
					Data: []byte("text"),
				},
			},
		},
	}

	req, err := internal.NewRequest(context.Background(), "POST", "/", op)
	testingx.Expect(t, err, testingx.BeNil[error]())
	testingx.Expect(t, req, testingutil.BeRequest(`
POST / HTTP/1.1
Content-Type: multipart/form-data; boundary=boundary1

--boundary1
Content-Disposition: form-data; name=a
Content-Length: 1
Content-Type: text/plain; charset=utf-8

s
--boundary1
Content-Disposition: form-data; name=filter
Content-Length: 2
Content-Type: text/plain; charset=utf-8

x1
--boundary1
Content-Disposition: form-data; name=filter
Content-Length: 2
Content-Type: text/plain; charset=utf-8

x2
--boundary1
Content-Disposition: form-data; filename=1.txt; name=file
Content-Type: text/plain

text
--boundary1
Content-Disposition: form-data; filename=2.txt; name=files
Content-Type: text/plain

text
--boundary1--
`))

	op2 := struct {
		Body Data `in:"body" mime:"multipart"`
	}{}

	err = internal.UnmarshalRequest(req, &op2)
	testingx.Expect(t, err, testingx.BeNil[error]())
	testingx.Expect(t, op2.Body, testingx.Equal(op.Body))
}

var (
	_ io.ReadCloser           = File{}
	_ internal.ReadCloserFrom = &File{}
)

type File struct {
	Name string
	Type string
	Data []byte
}

func (f *File) SetFilename(name string) {
	f.Name = name
}

func (f *File) SetContentType(ct string) {
	f.Type = ct
}

func (f *File) ReadFromCloser(r io.ReadCloser) (n int64, err error) {
	defer r.Close()

	bytes, err := io.ReadAll(r)
	f.Data = bytes
	return int64(len(bytes)), err
}

func (f File) IsZero() bool {
	return len(f.Data) == 0
}

func (f File) Read(p []byte) (n int, err error) {
	n = copy(p, f.Data)
	return n, io.EOF
}

func (f File) Close() error {
	return nil
}

func (f File) ContentType() string {
	return f.Type
}

func (f File) Filename() string {
	return f.Name
}
