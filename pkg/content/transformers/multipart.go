package transformers

import (
	"context"
	"errors"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"reflect"
	"strconv"

	"github.com/octohelm/courier/internal/jsonflags"
	"github.com/octohelm/courier/pkg/content/internal"
	validatorerrors "github.com/octohelm/courier/pkg/validator/errors"
)

func init() {
	internal.Register(&multipartTransformerProvider{})
}

type multipartTransformerProvider struct {
}

func (p *multipartTransformerProvider) Names() []string {
	return []string{
		"multipart/form-data", "multipart", "form-data",
	}
}

func (p *multipartTransformerProvider) Transformer() (internal.Transformer, error) {
	return &multipartTransformer{
		mediaType: p.Names()[0],
	}, nil
}

type multipartTransformer struct {
	mediaType string
}

func (p *multipartTransformer) MediaType() string {
	return p.mediaType
}

const (
	defaultMaxMemory = 32 << 20 // 32 MB
)

type withFilename interface {
	Filename() string
}

var withFilenameType = reflect.TypeFor[withFilename]()

func (p *multipartTransformer) ReadAs(ctx context.Context, r io.ReadCloser, v any) error {
	defer r.Close()

	header := http.Header{}
	if withHeader, ok := r.(internal.HeaderGetter); ok {
		header = withHeader.Header()
	}

	_, params, err := mime.ParseMediaType(header.Get("Content-Type"))
	if err != nil {
		return err
	}

	reader := multipart.NewReader(r, params["boundary"])
	form, err := reader.ReadForm(defaultMaxMemory)
	if err != nil {
		return err
	}

	rv, ok := v.(reflect.Value)
	if !ok {
		rv = reflect.ValueOf(v)
	}

	if rv.Kind() != reflect.Ptr {
		return errors.New("target must be ptr value")
	}

	pv := &internal.ParamValue{}
	pv.Value = rv.Elem()

	fields, err := jsonflags.Structs.StructFields(pv.Type())
	if err != nil {
		return err
	}

	var errs []error

	for sf := range fields.StructField() {
		if sf.Type.Implements(withFilenameType) || pv.CanMultiple(sf) && sf.Type.Elem().Implements(withFilenameType) {
			files := form.File[sf.Name]

			if len(files) == 0 {
				continue
			}

			readers := make([]io.ReadCloser, len(files))
			for i := range readers {
				file := files[i]
				f, err := file.Open()
				if err != nil {
					return err
				}
				readers[i] = internal.ReadCloseWithHeader(f, http.Header(file.Header))
			}

			if err := pv.UnmarshalReaders(ctx, sf, readers); err != nil {
				errs = append(errs, err)
			}
			continue
		}

		if err := pv.UnmarshalValues(ctx, sf, form.Value[sf.Name]); err != nil {
			errs = append(errs, err)
		}
	}

	return validatorerrors.Join(errs...)
}

func (mt *multipartTransformer) Prepare(ctx context.Context, v any) (internal.Content, error) {
	c := NewContent(mt.mediaType)
	c.ReadCloser = AsReaderCloser(ctx, func(w io.WriteCloser) func() error {
		mw := multipart.NewWriter(w)
		c.contentType = mw.FormDataContentType()

		return func() error {
			defer w.Close()

			rv, ok := v.(reflect.Value)
			if ok {
				v = rv.Interface()
			}

			for rv.Kind() == reflect.Pointer {
				rv = rv.Elem()
			}

			pv := &internal.ParamValue{}
			pv.Value = rv

			s, err := jsonflags.Structs.StructFields(pv.Type())
			if err != nil {
				return err
			}

			for sf := range s.StructField() {
				for sfv := range pv.Values(sf) {
					if sfv.IsZero() {
						if sf.Omitzero || sf.Omitempty {
							continue
						}
					}

					params := map[string]string{
						"name": sf.Name,
					}

					fv := sfv.Interface()
					if withFilename, ok := fv.(interface{ Filename() string }); ok {
						params["filename"] = withFilename.Filename()
					}

					header := textproto.MIMEHeader{}

					header.Set("Content-Disposition", mime.FormatMediaType("form-data", params))

					cw, err := internal.New(sfv.Type(), sf.Tag.Get("mime"), "marshal")
					if err != nil {
						return err
					}

					c, err := cw.Prepare(ctx, sfv)
					if err != nil {
						return err
					}

					if ct := c.GetContentType(); ct != "" {
						header.Set("Content-Type", ct)
					}

					if withContentType, ok := fv.(interface{ ContentType() string }); ok {
						header.Set("Content-Type", withContentType.ContentType())
					}

					if i := c.GetContentLength(); i > -1 {
						header.Set("Content-Length", strconv.FormatInt(i, 10))
					}

					p, err := mw.CreatePart(header)
					if err != nil {
						return err
					}

					if _, err := io.Copy(p, c); err != nil {
						return err
					}
				}
			}
			return mw.Close()
		}
	})

	return c, nil
}
