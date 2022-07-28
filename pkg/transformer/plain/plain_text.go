package plain

import (
	"context"
	"io"
	"net/textproto"

	"github.com/octohelm/courier/pkg/transformer/core"
	encodingx "github.com/octohelm/x/encoding"
	typesx "github.com/octohelm/x/types"
)

func init() {
	core.Register(&plainTextTranformer{})
}

type plainTextTranformer struct {
}

func (t *plainTextTranformer) String() string {
	return t.Names()[0]
}

func (*plainTextTranformer) Names() []string {
	return []string{"text/plain", "plain", "text", "txt"}
}

func (*plainTextTranformer) New(context.Context, typesx.Type) (core.Transformer, error) {
	return &plainTextTranformer{}, nil
}

func (t *plainTextTranformer) EncodeTo(ctx context.Context, w io.Writer, v any) error {
	core.WriteHeader(ctx, w, t.String(), map[string]string{
		"charset": "utf-8",
	})

	data, err := encodingx.MarshalText(v)
	if err != nil {
		return err
	}

	if _, err := w.Write(data); err != nil {
		return err
	}

	return nil
}

func (t *plainTextTranformer) DecodeFrom(ctx context.Context, r io.ReadCloser, v any, headers ...textproto.MIMEHeader) error {
	defer r.Close()

	switch x := r.(type) {
	case core.CanString:
		raw := x.String()
		if x, ok := v.(*string); ok {
			*x = raw
			return nil
		}
		return encodingx.UnmarshalText(v, []byte(raw))
	case core.CanInterface:
		if raw, ok := x.Interface().(string); ok {
			if x, ok := v.(*string); ok {
				*x = raw
				return nil
			}
			return encodingx.UnmarshalText(v, []byte(raw))
		}
	}
	data, err := io.ReadAll(r)
	if err != nil {
		return err
	}
	return encodingx.UnmarshalText(v, data)
}
