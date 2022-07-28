package urlencoded

import (
	"context"
	"io"
	"net/textproto"
	"net/url"
	"reflect"

	"github.com/octohelm/courier/pkg/transformer/core"
	verrors "github.com/octohelm/courier/pkg/validator"
	reflectx "github.com/octohelm/x/reflect"
	typesutil "github.com/octohelm/x/types"
	"github.com/pkg/errors"
)

func init() {
	core.Register(&urlEncodedTransformer{})
}

type urlEncodedTransformer struct {
	*core.FlattenParams
}

func (urlEncodedTransformer) Names() []string {
	return []string{"application/x-www-form-urlencoded", "form", "urlencoded", "url-encoded"}
}

func (urlEncodedTransformer) NamedByTag() string {
	return "name"
}

func (urlEncodedTransformer) New(ctx context.Context, typ typesutil.Type) (core.Transformer, error) {
	transformer := &urlEncodedTransformer{}

	typ = typesutil.Deref(typ)
	if typ.Kind() != reflect.Struct {
		return nil, errors.Errorf("content transformer `%s` should be used for struct type", transformer)
	}

	transformer.FlattenParams = &core.FlattenParams{}

	if err := transformer.FlattenParams.CollectParams(ctx, typ); err != nil {
		return nil, err
	}

	return transformer, nil
}

func (transformer *urlEncodedTransformer) EncodeTo(ctx context.Context, w io.Writer, v any) error {
	rv, ok := v.(reflect.Value)
	if !ok {
		rv = reflect.ValueOf(v)
	}
	rv = reflectx.Indirect(rv)

	values := url.Values{}
	errSet := verrors.NewErrorSet()

	for i := range transformer.Parameters {
		p := transformer.Parameters[i]

		if p.Transformer != nil {
			fieldValue := p.FieldValue(rv)
			stringBuilders := core.NewStringBuilders()
			if err := core.Wrap(p.Transformer, &p.TransformerOption.CommonOption).EncodeTo(ctx, stringBuilders, fieldValue); err != nil {
				errSet.AddErr(err, p.Name)
				continue
			}
			values[p.Name] = stringBuilders.StringSlice()
		}
	}

	if err := errSet.Err(); err != nil {
		return err
	}

	core.WriteHeader(ctx, w, transformer.Names()[0], map[string]string{
		"param": "value",
	})
	_, err := w.Write([]byte(values.Encode()))
	return err
}

func (transformer *urlEncodedTransformer) DecodeFrom(ctx context.Context, r io.ReadCloser, v any, headers ...textproto.MIMEHeader) error {
	defer r.Close()

	rv, ok := v.(reflect.Value)
	if !ok {
		rv = reflect.ValueOf(v)
	}

	if rv.Kind() != reflect.Ptr {
		return errors.New("decode target must be ptr value")
	}

	data, err := io.ReadAll(r)
	if err != nil {
		return err
	}

	values, err := url.ParseQuery(string(data))
	if err != nil {
		return err
	}

	es := verrors.NewErrorSet()

	for i := range transformer.Parameters {
		p := transformer.Parameters[i]

		fieldValues := values[p.Name]

		if len(fieldValues) == 0 {
			continue
		}

		if p.Transformer != nil {
			if err := core.Wrap(p.Transformer, &p.TransformerOption.CommonOption).DecodeFrom(ctx, core.NewStringReaders(fieldValues), p.FieldValue(rv).Addr()); err != nil {
				es.AddErr(err, p.Name)
				continue
			}
		}
	}

	return es.Err()
}
