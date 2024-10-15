package transformers

import (
	"context"
	"github.com/go-json-experiment/json"
	"io"
	"mime"
	"net/http"
	"reflect"

	"github.com/octohelm/courier/pkg/content/internal"
	"github.com/octohelm/courier/pkg/validator"
)

func init() {
	internal.Register(&jsonTransformerProvider{})
}

type jsonTransformerProvider struct {
}

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

func (p *jsonTransformer) ReadAs(ctx context.Context, r io.ReadCloser, v any) error {
	defer r.Close()

	rv, ok := v.(reflect.Value)
	if ok {
		v = rv.Interface()
	}

	return validator.UnmarshalRead(r, v)
}

func (p *jsonTransformer) PrepareWriter(headers http.Header, w io.Writer) internal.ContentWriter {
	if ct := headers.Get("Content-Type"); ct == "" {
		headers.Set("Content-Type", mime.FormatMediaType(p.mediaType, map[string]string{
			"charset": "utf-8",
		}))
	}

	return &jsonWriter{
		Writer: w,
	}
}

type jsonWriter struct {
	io.Writer
}

func (w *jsonWriter) Send(ctx context.Context, v any) error {
	rv, ok := v.(reflect.Value)
	if ok {
		v = rv.Interface()
	}

	if direct, ok := v.(json.MarshalerV1); ok {
		raw, err := direct.MarshalJSON()
		if err != nil {
			return err
		}
		_, err = w.Write(raw)
		return err
	}

	return validator.MarshalWrite(w, v)
}
