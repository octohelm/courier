package octet

import (
	"context"
	"io"
	"net/http"
	"net/textproto"
	"reflect"

	"github.com/octohelm/courier/pkg/courierhttp"

	"github.com/octohelm/courier/pkg/transformer/core"
	typesx "github.com/octohelm/x/types"
)

func init() {
	core.Register(&octetStreamTransformer{})
}

type octetStreamTransformer struct {
}

func (t *octetStreamTransformer) String() string {
	return t.Names()[0]
}

func (*octetStreamTransformer) Names() []string {
	return []string{"application/octet-stream", "octet-stream", "octet"}
}

func (*octetStreamTransformer) New(context.Context, typesx.Type) (core.Transformer, error) {
	return &octetStreamTransformer{}, nil
}

func (t *octetStreamTransformer) EncodeTo(ctx context.Context, w io.Writer, v any) error {
	rv, ok := v.(reflect.Value)
	if ok {
		v = rv.Interface()
	}

	switch x := v.(type) {
	case courierhttp.FileHeader:
		if rw, ok := w.(interface{ Header() http.Header }); ok {
			for k, hv := range x.Header() {
				rw.Header()[k] = hv
			}
		}
	}

	if x, ok := v.(io.Reader); ok {
		core.WriteHeader(ctx, w, t.Names()[0], nil)

		if rc, ok := x.(io.ReadCloser); ok {
			defer rc.Close()
		}

		if _, err := io.Copy(w, x); err != nil {
			return err
		}
	}

	return nil
}

func (*octetStreamTransformer) DecodeFrom(ctx context.Context, r io.ReadCloser, v any, headers ...textproto.MIMEHeader) error {
	rv, ok := v.(reflect.Value)
	if !ok {
		rv = reflect.ValueOf(v)
	}

	switch x := rv.Interface().(type) {
	case *io.ReadCloser:
		*x = r
	case io.Writer:
		defer r.Close()
		if _, err := io.Copy(x, r); err != nil {
			return err
		}
	}
	return nil
}
