package transformers

import (
	"bytes"
	"context"
	"io"
	"mime"
	"reflect"

	"github.com/octohelm/courier/internal/jsonflags"

	"github.com/go-json-experiment/json/jsontext"
	"github.com/octohelm/courier/pkg/content/internal"
	"github.com/octohelm/courier/pkg/validator"
)

func init() {
	internal.Register(&textTransformerProvider{})
}

type textTransformerProvider struct{}

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

func (p *textTransformer) ReadAs(ctx context.Context, r io.ReadCloser, vv any) error {
	defer r.Close()

	v := jsonflags.Unwrap(vv)

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
		return validator.Unmarshal(raw, vv)
	}
}

func (p *textTransformer) Prepare(ctx context.Context, v any) (internal.Content, error) {
	c := NewContent(mime.FormatMediaType(p.mediaType, map[string]string{
		"charset": "utf-8",
	}))

	rv, ok := v.(reflect.Value)
	if ok {
		v = rv.Interface()
	}

	switch x := v.(type) {
	case []byte:
		c.SetContentLength(int64(len(x)))
		c.ReadCloser = io.NopCloser(bytes.NewBuffer(x))
		return c, nil
	case string:
		c.SetContentLength(int64(len(x)))
		c.ReadCloser = io.NopCloser(bytes.NewBufferString(x))
		return c, nil
	default:
		data, err := validator.Marshal(v)
		if err != nil {
			return nil, err
		}
		raw, err := jsontext.AppendUnquote(nil, data)
		if err != nil {
			return nil, err
		}
		c.SetContentLength(int64(len(raw)))
		c.ReadCloser = io.NopCloser(bytes.NewBuffer(raw))
		return c, nil
	}
}
