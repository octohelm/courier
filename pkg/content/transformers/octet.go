package transformers

import (
	"bytes"
	"context"
	"github.com/octohelm/courier/internal/jsonflags"
	"io"
	"mime"
	"reflect"

	"github.com/go-json-experiment/json"
	"github.com/go-json-experiment/json/jsontext"
	"github.com/octohelm/courier/pkg/content/internal"
	"github.com/octohelm/courier/pkg/validator"
)

func init() {
	internal.Register(&octecTransformerProvider{})
}

type octecTransformerProvider struct{}

func (p *octecTransformerProvider) Names() []string {
	return []string{
		"application/octet-stream", "octet-stream", "octet",
	}
}

func (p *octecTransformerProvider) Transformer() (internal.Transformer, error) {
	return &octecTransformer{
		mediaType: p.Names()[0],
	}, nil
}

type octecTransformer struct {
	mediaType string
}

func (p *octecTransformer) MediaType() string {
	return p.mediaType
}

func (p *octecTransformer) ReadAs(ctx context.Context, r io.ReadCloser, vv any) error {
	v := jsonflags.Unwrap(vv)

	rv, ok := v.(reflect.Value)
	if ok {
		v = rv.Interface()
	}

	if x, ok := v.(*io.ReadCloser); ok {
		*x = r
		return nil
	}

	header, ok := r.(internal.HeaderGetter)
	if ok {
		if x, ok := v.(internal.FilenameSetter); ok {
			_, params, err := mime.ParseMediaType(header.Header().Get("Content-Disposition"))
			if err == nil {
				x.SetFilename(params["filename"])
			}
		}

		if x, ok := v.(internal.ContentTypeSetter); ok {
			x.SetContentType(header.Header().Get("Content-Type"))
		}
	}

	if x, ok := v.(internal.ReadCloserFrom); ok {
		_, err := x.ReadFromCloser(r)
		return err
	}

	defer r.Close()

	switch x := v.(type) {
	case io.ReaderFrom:
		if _, err := x.ReadFrom(r); err != nil {
			return err
		}
		return nil
	case io.Writer:
		if _, err := io.Copy(x, r); err != nil {
			return err
		}
		return nil
	case *[]byte:
		data, err := io.ReadAll(r)
		if err != nil {
			return err
		}
		*x = data
		return nil
	default:
		data, err := io.ReadAll(r)
		if err != nil {
			return err
		}
		raw, err := jsontext.AppendQuote(nil, data)
		if err != nil {
			return err
		}
		return json.Unmarshal(raw, v)
	}
}

func (p *octecTransformer) Prepare(ctx context.Context, v any) (internal.Content, error) {
	c := NewContent(p.mediaType)

	rv, ok := v.(reflect.Value)
	if ok {
		v = rv.Interface()
	}

	if v == nil {
		c.SetContentLength(0)
		c.ReadCloser = io.NopCloser(bytes.NewBuffer(nil))
		return c, nil
	}

	switch x := v.(type) {
	case io.ReadCloser:
		c.ReadCloser = x
		return c, nil
	case io.Reader:
		c.ReadCloser = io.NopCloser(x)
		return c, nil
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
