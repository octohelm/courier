package json

import (
	"context"
	"github.com/go-json-experiment/json"
	"github.com/go-json-experiment/json/jsontext"
	transformer "github.com/octohelm/courier/pkg/transformer/core"
	typesutil "github.com/octohelm/x/types"
	"io"
	"net/textproto"
	"reflect"
)

func init() {
	transformer.Register(&jsonTransformer{})
}

type jsonTransformer struct {
}

func (*jsonTransformer) Names() []string {
	return []string{"application/json", "json"}
}

func (*jsonTransformer) NamedByTag() string {
	return "json"
}

func (transformer *jsonTransformer) String() string {
	return transformer.Names()[0]
}

func (*jsonTransformer) New(context.Context, typesutil.Type) (transformer.Transformer, error) {
	return &jsonTransformer{}, nil
}

func (t *jsonTransformer) EncodeTo(ctx context.Context, w io.Writer, v any) error {
	if rv, ok := v.(reflect.Value); ok {
		v = rv.Interface()
	}

	transformer.WriteHeader(ctx, w, t.String(), map[string]string{
		"charset": "utf-8",
	})

	if m, ok := v.(json.MarshalerV1); ok {
		data, err := m.MarshalJSON()
		if err != nil {
			return err
		}
		_, err = w.Write(data)
		return err
	}

	return json.MarshalEncode(jsontext.NewEncoder(w), v)
}

func (*jsonTransformer) DecodeFrom(ctx context.Context, r io.ReadCloser, v any, headers ...textproto.MIMEHeader) error {
	defer r.Close()

	if rv, ok := v.(reflect.Value); ok {
		if rv.Kind() != reflect.Ptr && rv.CanAddr() {
			rv = rv.Addr()
		}
		v = rv.Interface()
	}

	dec := jsontext.NewDecoder(r)
	if err := json.UnmarshalDecode(dec, v); err != nil {
		return err
	}
	return nil
}
