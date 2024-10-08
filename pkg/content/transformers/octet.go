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
	internal.Register(&octecTransformerProvider{})
}

type octecTransformerProvider struct {
}

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

func (p *octecTransformer) ReadAs(ctx context.Context, r io.ReadCloser, v any) error {
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

func (p *octecTransformer) PrepareWriter(headers http.Header, w io.Writer) internal.ContentWriter {
	if ct := headers.Get("Content-Type"); ct == "" {
		headers.Set("Content-Type", p.mediaType)
	}

	return &octecWriter{
		Writer: w,
	}
}

type octecWriter struct {
	io.Writer
}

func (w *octecWriter) Send(ctx context.Context, v any) error {
	rv, ok := v.(reflect.Value)
	if ok {
		v = rv.Interface()
	}

	switch x := v.(type) {
	case io.ReadCloser:
		defer x.Close()

		if _, err := io.Copy(w, x); err != nil {
			return err
		}
		return nil
	case io.Reader:
		if _, err := io.Copy(w, x); err != nil {
			return err
		}
		return nil
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
		if _, err := w.Write(raw); err != nil {
			return err
		}
		return nil
	}
}
