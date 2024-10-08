package transformers

import (
	"context"
	"github.com/octohelm/courier/pkg/validator"
	"io"
	"mime"
	"net/http"
	"reflect"

	"github.com/go-json-experiment/json"
	"github.com/go-json-experiment/json/jsontext"
	"github.com/octohelm/courier/pkg/content/internal"
)

func init() {
	internal.Register(&textTransformerProvider{})
}

type textTransformerProvider struct {
}

func (p *textTransformerProvider) Names() []string {
	return []string{
		"text/plain", "plain", "text", "txt",
	}
}

func (p *textTransformerProvider) Transformer() (internal.Transformer, error) {
	return &textTransformer{
		mediaType: p.Names()[0],
	}, nil
}

type textTransformer struct {
	mediaType string
}

func (p *textTransformer) MediaType() string {
	return p.mediaType
}

func (p *textTransformer) ReadAs(ctx context.Context, r io.ReadCloser, v any) error {
	defer r.Close()

	rv, ok := v.(reflect.Value)
	if ok {
		v = rv.Interface()
	}

	data, err := io.ReadAll(r)
	if err != nil {
		return err
	}

	switch x := v.(type) {
	case *[]byte:
		*x = data
		return nil
	default:
		raw, err := jsontext.AppendQuote(nil, data)
		if err != nil {
			return err
		}
		return json.Unmarshal(raw, v)
	}
}

func (p *textTransformer) PrepareWriter(headers http.Header, w io.Writer) internal.ContentWriter {
	if ct := headers.Get("Content-Type"); ct == "" {
		headers.Set("Content-Type", mime.FormatMediaType(p.mediaType, map[string]string{
			"charset": "utf-8",
		}))
	}

	return &textWriter{
		Writer: w,
	}
}

type textWriter struct {
	io.Writer
}

func (w *textWriter) Send(ctx context.Context, v any) error {
	rv, ok := v.(reflect.Value)
	if ok {
		v = rv.Interface()
	}

	switch x := v.(type) {
	case []byte:
		if _, err := w.Write(x); err != nil {
			return err
		}
		return nil
	case string:
		if _, err := w.Write([]byte(x)); err != nil {
			return err
		}
		return nil
	default:
		data, err := validator.Marshal(v)
		if err != nil {
			return err
		}
		raw, err := jsontext.AppendUnquote(nil, data)
		if err != nil {
			return err
		}
		if _, err := w.Writer.Write(raw); err != nil {
			return err
		}
		return nil
	}
}
