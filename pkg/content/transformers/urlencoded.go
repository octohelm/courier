package transformers

import (
	"context"
	"io"
	"mime"
	"net/url"
	"reflect"

	"github.com/go-json-experiment/json/jsontext"
	"github.com/octohelm/courier/pkg/content/internal"
	"github.com/octohelm/courier/pkg/validator"
)

func init() {
	internal.Register(&urlencodedTransformerProvider{})
}

type urlencodedTransformerProvider struct{}

func (p *urlencodedTransformerProvider) Names() []string {
	return []string{
		"application/x-www-form-urlencoded", "form", "urlencoded", "url-encoded",
	}
}

func (p *urlencodedTransformerProvider) Transformer() (internal.Transformer, error) {
	return &urlencodedTransformer{
		mediaType: p.Names()[0],
	}, nil
}

type urlencodedTransformer struct {
	mediaType string
}

func (p *urlencodedTransformer) MediaType() string {
	return p.mediaType
}

func (p *urlencodedTransformer) ReadAs(ctx context.Context, r io.ReadCloser, v any) error {
	defer r.Close()

	return internal.Pipe(
		func(w io.Writer) error {
			data, err := io.ReadAll(r)
			if err != nil {
				return err
			}

			values, err := url.ParseQuery(string(data))
			if err != nil {
				return err
			}

			if err = validator.MarshalWrite(w, values); err != nil {
				return err
			}
			return nil
		},
		func(r io.Reader) error {
			return validator.UnmarshalRead(r, v)
		},
	)
}

func (p *urlencodedTransformer) Prepare(ctx context.Context, v any) (internal.Content, error) {
	c := NewContent(mime.FormatMediaType(p.mediaType, map[string]string{
		"param": "value",
	}))

	rv, ok := v.(reflect.Value)
	if ok {
		v = rv.Interface()
	}

	c.ReadCloser = AsReaderCloser(ctx, func(w io.WriteCloser) func() error {
		uw := &urlencodedWriter{Writer: w}

		return func() error {
			defer w.Close()

			return internal.Pipe(
				func(w io.Writer) error {
					return validator.MarshalWrite(w, v)
				},
				func(r io.Reader) error {
					return uw.urlencodedWrite(jsontext.NewDecoder(r))
				},
			)
		}
	})

	return c, nil
}

type urlencodedWriter struct {
	io.Writer
}

// { "a": 1 } => a=1
func (w *urlencodedWriter) urlencodedWrite(dec *jsontext.Decoder) error {
	tok, err := dec.ReadToken()
	if err != nil {
		return err
	}
	k := tok.Kind()
	switch k {
	case '{':
		i := 0
		for dec.PeekKind() != '}' {
			keyTok, err := dec.ReadToken()
			if err != nil {
				return err
			}
			key := keyTok.String()

			t := dec.PeekKind()
			if t == '[' {
				if _, err := dec.ReadToken(); err != nil {
					return err
				}

				for dec.PeekKind() != ']' {
					value, err := dec.ReadValue()
					if err != nil {
						return err
					}

					written, err := w.paramValueWrite(i, key, value)
					if err != nil {
						return err
					}
					if written {
						i++
					}
				}

				if _, err := dec.ReadToken(); err != nil {
					return err
				}
			} else {
				value, err := dec.ReadValue()
				if err != nil {
					return err
				}

				written, err := w.paramValueWrite(i, key, value)
				if err != nil {
					return err
				}
				if written {
					i++
				}
			}
		}

		// read }
		if _, err := dec.ReadToken(); err != nil {
			return err
		}

		return nil
	}
	return nil
}

func (w *urlencodedWriter) paramValueWrite(i int, key string, value jsontext.Value) (bool, error) {
	if i > 0 {
		if _, err := w.Write([]byte("&")); err != nil {
			return false, err
		}
	}

	switch value.Kind() {
	case 'n':

	case '"':
		unquote, err := jsontext.AppendUnquote(nil, value)
		if err != nil {
			return false, err
		}
		if _, err := w.Write([]byte(url.QueryEscape(key) + "=")); err != nil {
			return false, err
		}
		if _, err := w.Write([]byte(url.QueryEscape(string(unquote)))); err != nil {
			return false, err
		}
		return true, nil
	default:
		if _, err := w.Write([]byte(url.QueryEscape(key) + "=")); err != nil {
			return false, err
		}
		if _, err := w.Write([]byte(url.QueryEscape(string(value)))); err != nil {
			return false, err
		}
	}
	return false, nil
}
