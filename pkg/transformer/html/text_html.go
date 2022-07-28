package html

import (
	"context"
	"io"
	"net/textproto"
	"reflect"

	"github.com/octohelm/courier/pkg/transformer/core"

	transformer "github.com/octohelm/courier/pkg/transformer/core"
	encodingx "github.com/octohelm/x/encoding"
	typesutil "github.com/octohelm/x/types"
)

func init() {
	transformer.Register(&textHtml{})
}

type textHtml struct {
}

func (*textHtml) NamedByTag() string {
	return ""
}

func (t *textHtml) String() string {
	return t.Names()[0]
}

func (*textHtml) Names() []string {
	return []string{"text/html", "html"}
}

func (*textHtml) New(context.Context, typesutil.Type) (transformer.Transformer, error) {
	return &textHtml{}, nil
}

func (t *textHtml) EncodeTo(ctx context.Context, w io.Writer, v any) error {
	rv, ok := v.(reflect.Value)
	if !ok {
		rv = reflect.ValueOf(v)
	}

	core.WriteHeader(ctx, w, t.String(), map[string]string{
		"charset": "utf-8",
	})

	data, err := encodingx.MarshalText(rv)
	if err != nil {
		return err
	}
	if _, err := w.Write(data); err != nil {
		return err
	}
	return nil
}

func (*textHtml) DecodeFrom(ctx context.Context, r io.ReadCloser, v any, headers ...textproto.MIMEHeader) error {
	defer r.Close()
	rv, ok := v.(reflect.Value)
	if !ok {
		rv = reflect.ValueOf(v)
	}
	data, err := io.ReadAll(r)
	if err != nil {
		return err
	}
	return encodingx.UnmarshalText(rv, data)
}
