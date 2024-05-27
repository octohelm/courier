package core

import (
	"context"
	"io"
	"mime"
	"net/http"
	"net/textproto"
	"net/url"
	"reflect"
	"sync"

	contextx "github.com/octohelm/x/context"
	typesx "github.com/octohelm/x/types"
	"github.com/pkg/errors"
)

type Mgr interface {
	NewTransformer(context.Context, typesx.Type, Option) (Transformer, error)
	GetTransformerNames(name string) []string
}

type contextKeyMgr struct{}

func ContextWithMgr(ctx context.Context, mgr Mgr) context.Context {
	return contextx.WithValue(ctx, contextKeyMgr{}, mgr)
}

func MgrFromContext(ctx context.Context) Mgr {
	if mgr, ok := ctx.Value(contextKeyMgr{}).(Mgr); ok {
		return mgr
	}
	return mgrDefault
}

func NewTransformer(ctx context.Context, tpe typesx.Type, opt Option) (Transformer, error) {
	return MgrFromContext(ctx).NewTransformer(ctx, tpe, opt)
}

var mgrDefault = &transformerFactory{}

func Register(transformers ...Transformer) {
	mgrDefault.Register(transformers...)
}

type Transformer interface {
	// Names should include name or alias of transformer
	// prefer using some keyword about content-type
	// first must validate content-type
	Names() []string
	// New will create transformer new transformer instance by type
	// in this step will to check transformer is valid for type
	New(context.Context, typesx.Type) (Transformer, error)

	TransformerEncoder
	TransformerDecoder
}

type TransformerEncoder interface {
	// EncodeTo
	// if w implement interface { Header() http.Header }
	// Content-Type will be set
	EncodeTo(ctx context.Context, w io.Writer, v any) (err error)
}

type TransformerDecoder interface {
	// DecodeFrom
	// will unmarshal data from read into some struct
	DecodeFrom(ctx context.Context, r io.ReadCloser, v any, headers ...textproto.MIMEHeader) error
}

type Option struct {
	Name   string
	MIME   string
	Strict bool
	CommonOption
}

type CommonOption struct {
	// when enable
	// should ignore value when value is empty
	Omitempty bool
	Explode   bool
}

func (op Option) String() string {
	values := url.Values{}

	if op.Name != "" {
		values.Add("OrgName", op.Name)
	}

	if op.MIME != "" {
		values.Add("MIME", op.MIME)
	}

	if op.Omitempty {
		values.Add("Omitempty", "true")
	}

	if op.Explode {
		values.Add("Explode", "true")
	}

	return values.Encode()
}

type transformerFactory struct {
	transformerSet map[string]Transformer
	cache          sync.Map
}

func (c *transformerFactory) Register(transformers ...Transformer) {
	if c.transformerSet == nil {
		c.transformerSet = map[string]Transformer{}
	}
	for i := range transformers {
		transformer := transformers[i]
		for _, name := range transformer.Names() {
			c.transformerSet[name] = transformer
		}
	}
}

func (c *transformerFactory) GetTransformerNames(name string) []string {
	if ct, ok := c.transformerSet[name]; ok {
		return ct.Names()
	}
	return nil
}

func (c *transformerFactory) NewTransformer(ctx context.Context, typ typesx.Type, opt Option) (Transformer, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	key := typesx.FullTypeName(typ) + opt.String()

	if v, ok := c.cache.Load(key); ok {
		return v.(Transformer), nil
	}

	if opt.MIME == "" {
		indirectType := typesx.Deref(typ)

		// io.ReadCloser
		if indirectType.PkgPath() == "io" && indirectType.Name() == "ReadCloser" {
			opt.MIME = "octet-stream"
		} else {
			switch indirectType.Kind() {
			case reflect.Slice:
				if indirectType.Elem().PkgPath() == "" && indirectType.Elem().Kind() == reflect.Uint8 {
					// bytes
					opt.MIME = "plain"
				} else {
					opt.MIME = "json"
				}
			case reflect.Struct:
				opt.MIME = "json"
			case reflect.Map, reflect.Array:
				opt.MIME = "json"
			default:
				opt.MIME = "plain"
			}

			if _, ok := typesx.EncodingTextMarshalerTypeReplacer(typ); ok {
				opt.MIME = "plain"
			}
		}
	}

	if ct, ok := c.transformerSet[opt.MIME]; ok {
		contentTransformer, err := ct.New(ContextWithMgr(ctx, c), typ)
		if err != nil {
			return nil, err
		}
		c.cache.Store(key, contentTransformer)
		return contentTransformer, nil
	}

	return nil, errors.Errorf("fmt %s is not supported for content transformer", key)
}

func WriteHeader(ctx context.Context, w io.Writer, contentType string, param map[string]string) {
	if rw, ok := w.(interface{ Header() http.Header }); ok {
		// only set content when empty
		if ct := rw.Header().Get("Content-Type"); ct == "" {
			if len(param) == 0 {
				rw.Header().Set("Content-Type", contentType)
			} else {
				rw.Header().Set("Content-Type", mime.FormatMediaType(contentType, param))
			}
		}
	}

	if rw, ok := w.(http.ResponseWriter); ok {
		rw.WriteHeader(StatusCodeFromContext(ctx))
	}
}

type contextKeyStatusCode struct{}

func ContextWithStatusCode(ctx context.Context, statusCode int) context.Context {
	return context.WithValue(ctx, contextKeyStatusCode{}, statusCode)
}

func StatusCodeFromContext(ctx context.Context) int {
	if statusCode, ok := ctx.Value(contextKeyStatusCode{}).(int); ok {
		return statusCode
	}
	return http.StatusOK
}
