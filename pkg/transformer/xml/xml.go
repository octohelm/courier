package xml

import (
	"context"
	"encoding/xml"
	"io"
	"net/textproto"
	"reflect"

	transformer "github.com/octohelm/courier/pkg/transformer/core"

	typesutil "github.com/octohelm/x/types"
)

func init() {
	transformer.Register(&xmlTransformer{})
}

type xmlTransformer struct {
}

func (*xmlTransformer) Names() []string {
	return []string{"application/xml", "xml"}
}

func (t *xmlTransformer) String() string {
	return t.Names()[0]
}

func (*xmlTransformer) NamedByTag() string {
	return "xml"
}

func (*xmlTransformer) New(context.Context, typesutil.Type) (transformer.Transformer, error) {
	return &xmlTransformer{}, nil
}

func (t *xmlTransformer) EncodeTo(ctx context.Context, w io.Writer, v any) error {
	if rv, ok := v.(reflect.Value); ok {
		v = rv.Interface()
	}

	transformer.WriteHeader(ctx, w, t.String(), map[string]string{
		"charset": "utf-8",
	})

	return xml.NewEncoder(w).Encode(v)
}

func (*xmlTransformer) DecodeFrom(ctx context.Context, r io.ReadCloser, v any, headers ...textproto.MIMEHeader) error {
	defer r.Close()
	if rv, ok := v.(reflect.Value); ok {
		if rv.Kind() != reflect.Ptr && rv.CanAddr() {
			rv = rv.Addr()
		}
		v = rv.Interface()
	}
	d := xml.NewDecoder(r)
	err := d.Decode(v)
	if err != nil {
		// todo resolve field path by InputOffset()
		return err
	}
	return nil
}
