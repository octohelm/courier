package transformers

import (
	"bytes"
	"context"
	"io"
	"mime"
	"reflect"

	"github.com/octohelm/courier/internal/jsonflags"

	"github.com/go-json-experiment/json"
	"github.com/octohelm/courier/pkg/content/internal"
	"github.com/octohelm/courier/pkg/validator"
)

func init() {
	internal.Register(&jsonTransformerProvider{})
}

type jsonTransformerProvider struct{}

func (p *jsonTransformerProvider) Names() []string {
	return []string{
		"application/json", "json",
	}
}

func (p *jsonTransformerProvider) Transformer() (internal.Transformer, error) {
	return &jsonTransformer{
		mediaType: p.Names()[0],
	}, nil
}

type jsonTransformer struct {
	mediaType string
}

func (p *jsonTransformer) MediaType() string {
	return p.mediaType
}

func (p *jsonTransformer) ReadAs(ctx context.Context, r io.ReadCloser, i any) error {
	v := jsonflags.Unwrap(i)
	defer r.Close()

	rv, ok := v.(reflect.Value)
	if ok {
		v = rv.Interface()
	}

	if direct, ok := v.(json.Unmarshaler); ok {
		// avoid trim \n
		raw, err := io.ReadAll(r)
		if err != nil {
			return err
		}
		return direct.UnmarshalJSON(raw)
	}

	return validator.UnmarshalRead(r, i)
}

func (p *jsonTransformer) Prepare(ctx context.Context, v any) (internal.Content, error) {
	c := NewContent(mime.FormatMediaType(p.mediaType, map[string]string{
		"charset": "utf-8",
	}))

	rv, ok := v.(reflect.Value)
	if ok {
		v = rv.Interface()
	}

	if direct, ok := v.(json.Marshaler); ok {
		// avoid trim \n
		raw, err := direct.MarshalJSON()
		if err != nil {
			return nil, err
		}

		c.SetContentLength(int64(len(raw)))
		c.ReadCloser = io.NopCloser(bytes.NewBuffer(raw))

		return c, nil
	}

	c.ReadCloser = AsReaderCloser(ctx, func(w io.WriteCloser) func() error {
		return func() error {
			defer w.Close()

			return validator.MarshalWrite(w, v)
		}
	})

	return c, nil
}
