package multipart

import (
	"context"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"reflect"
	"strconv"

	"github.com/octohelm/courier/pkg/courierhttp"

	"github.com/octohelm/courier/pkg/transformer/core"
	verrors "github.com/octohelm/courier/pkg/validator"
	typesutil "github.com/octohelm/x/types"
	"github.com/pkg/errors"
)

func init() {
	core.Register(&multipartTransformer{})
}

type multipartTransformer struct {
	*core.FlattenParams
}

func (*multipartTransformer) Names() []string {
	return []string{"multipart/form-data", "multipart", "form-data"}
}

func (*multipartTransformer) NamedByTag() string {
	return "name"
}

func (transformer *multipartTransformer) String() string {
	return transformer.Names()[0]
}

func (*multipartTransformer) New(ctx context.Context, typ typesutil.Type) (core.Transformer, error) {
	transformer := &multipartTransformer{}

	typ = typesutil.Deref(typ)
	if typ.Kind() != reflect.Struct {
		return nil, errors.Errorf("content transformer `%s` should be used for struct type", transformer)
	}

	transformer.FlattenParams = &core.FlattenParams{}

	if err := transformer.FlattenParams.CollectParams(ctx, typ); err != nil {
		return nil, err
	}

	return transformer, nil
}

func (transformer *multipartTransformer) EncodeTo(ctx context.Context, w io.Writer, v any) error {
	rv, ok := v.(reflect.Value)
	if !ok {
		rv = reflect.ValueOf(v)
	}

	multipartWriter := multipart.NewWriter(w)

	core.WriteHeader(ctx, w, multipartWriter.FormDataContentType(), nil)

	errSet := verrors.NewErrorSet()

	for i := range transformer.Parameters {
		p := transformer.Parameters[i]

		fieldValue := p.FieldValue(rv)

		if p.Transformer != nil {
			st := core.Wrap(p.Transformer, &p.TransformerOption.CommonOption)

			partWriter := NewFormPartWriter(func(header textproto.MIMEHeader) (io.Writer, error) {
				paramFilename := ""
				if v := header.Get("Content-Disposition"); v != "" {
					_, disposition, err := mime.ParseMediaType(v)
					if err == nil {
						if f, ok := disposition["filename"]; ok {
							paramFilename = fmt.Sprintf("; filename=%s", strconv.Quote(f))
						}
					}
				}
				// always overwrite name
				header.Set("Content-Disposition", fmt.Sprintf("form-data; name=%s%s", strconv.Quote(p.Name), paramFilename))
				return multipartWriter.CreatePart(header)
			})

			if err := st.EncodeTo(ctx, partWriter, fieldValue); err != nil {
				errSet.AddErr(err, p.Name)
				continue
			}
		}
	}

	return multipartWriter.Close()
}

const (
	defaultMaxMemory = 32 << 20 // 32 MB
)

func (transformer *multipartTransformer) DecodeFrom(ctx context.Context, r io.ReadCloser, v any, headers ...textproto.MIMEHeader) error {
	defer r.Close()

	rv, ok := v.(reflect.Value)
	if !ok {
		rv = reflect.ValueOf(v)
	}

	header := core.MIMEHeader(headers...)
	_, params, err := mime.ParseMediaType(header.Get("Content-Type"))
	if err != nil {
		return err
	}

	reader := multipart.NewReader(r, params["boundary"])
	form, err := reader.ReadForm(defaultMaxMemory)
	if err != nil {
		return err
	}

	errSet := verrors.NewErrorSet()

	for i := range transformer.Parameters {
		p := transformer.Parameters[i]

		if p.Transformer != nil {
			st := core.Wrap(p.Transformer, &p.TransformerOption.CommonOption)

			if files, ok := form.File[p.Name]; ok {
				readers := wrapFileHeaderReaders(files)
				if err := st.DecodeFrom(ctx, readers, p.FieldValue(rv).Addr()); err != nil {
					errSet.AddErr(err, p.Name)
				}
				continue
			}

			if fieldValues, ok := form.Value[p.Name]; ok {
				readers := core.NewStringReaders(fieldValues)

				if err := st.DecodeFrom(ctx, readers, p.FieldValue(rv).Addr()); err != nil {
					errSet.AddErr(err, p.Name)
				}
			}
		}
	}

	return nil
}

func WithFilename(filename string) OptionFunc {
	return func(params map[string]string) {
		params["filename"] = filename
	}
}

func WithName(name string) OptionFunc {
	return func(params map[string]string) {
		params["name"] = name
	}
}

type OptionFunc = func(params map[string]string)

func WrapFileHeader(r io.ReadCloser, fns ...OptionFunc) courierhttp.FileHeader {
	params := map[string]string{}

	for i := range fns {
		fns[i](params)
	}

	return &fileHeader{
		params:     params,
		ReadCloser: r,
	}
}

type fileHeader struct {
	params map[string]string
	io.ReadCloser
}

func (f *fileHeader) Filename() string {
	if n, ok := f.params["filename"]; ok {
		return n
	}
	return ""
}

func (f *fileHeader) Header() http.Header {
	return http.Header{
		"Content-Disposition": {
			mime.FormatMediaType("form-data", f.params),
		},
	}
}

func NewFormPartWriter(createPartWriter func(header textproto.MIMEHeader) (io.Writer, error)) *FormPartWriter {
	return &FormPartWriter{
		createPartWriter: createPartWriter,
		header:           http.Header{},
	}
}

type FormPartWriter struct {
	createPartWriter func(header textproto.MIMEHeader) (io.Writer, error)
	partWriter       io.Writer
	header           http.Header
}

func (w *FormPartWriter) NextWriter() io.Writer {
	return NewFormPartWriter(w.createPartWriter)
}

func (w *FormPartWriter) Header() http.Header {
	return w.header
}

func (w *FormPartWriter) Write(p []byte) (n int, err error) {
	if w.partWriter == nil {
		w.partWriter, err = w.createPartWriter(textproto.MIMEHeader(w.header))
		if err != nil {
			return -1, err
		}
	}
	return w.partWriter.Write(p)
}

func wrapFileHeaderReaders(fileHeaders []*multipart.FileHeader) io.ReadCloser {
	bs := make([]io.ReadCloser, len(fileHeaders))
	for i := range fileHeaders {
		bs[i] = &multipartFileHeaderReader{
			v: fileHeaders[i],
		}
	}
	return &core.StringReaders{
		Readers: bs,
	}
}

type multipartFileHeaderReader struct {
	v      *multipart.FileHeader
	opened multipart.File
}

func (f *multipartFileHeaderReader) Filename() string {
	return f.v.Filename
}

func (f *multipartFileHeaderReader) Header() http.Header {
	return http.Header(f.v.Header)
}

func (f *multipartFileHeaderReader) Len() int64 {
	return f.v.Size
}

func (f *multipartFileHeaderReader) Read(p []byte) (int, error) {
	if f.opened == nil {
		file, err := f.v.Open()
		if err != nil {
			return -1, err
		}
		f.opened = file
	}
	return f.opened.Read(p)
}

func (f *multipartFileHeaderReader) Close() error {
	if f.opened == nil {
		return f.opened.Close()
	}
	return nil
}
