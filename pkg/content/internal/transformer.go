package internal

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"sync"

	"github.com/go-courier/logr"
)

type TransformerProvider interface {
	Names() []string
	Transformer() (Transformer, error)
}

type Transformer interface {
	MediaType() string

	ContentReader
	ContentWriterProvider
}

type ContentReader interface {
	ReadAs(ctx context.Context, r io.ReadCloser, v any) error
}

type ContentWriterProvider interface {
	PrepareWriter(headers http.Header, w io.Writer) ContentWriter
}

type ContentWriter interface {
	Send(ctx context.Context, v any) error
}

func AsReadCloser(ctx context.Context, cw ContentWriterProvider, v any, headers http.Header) (io.ReadCloser, error) {
	pr, pw := io.Pipe()

	w := cw.PrepareWriter(headers, pw)

	go func() {
		defer pw.Close()

		if err := w.Send(ctx, v); err != nil {
			logr.FromContext(ctx).Error(err)
		}
	}()

	return pr, nil
}

var defaultTransformers = &transformers{
	providers: map[string]TransformerProvider{},
}

func Register(creator TransformerProvider) {
	for _, name := range creator.Names() {
		defaultTransformers.providers[name] = creator
	}
}

func New(typ reflect.Type, mediaTypeOrAlias string, action string) (Transformer, error) {
	for typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}
	return defaultTransformers.New(transformOption{action: action, tpe: typ, mediaType: mediaTypeOrAlias})
}

type transformers struct {
	providers map[string]TransformerProvider
	// map[reflect.Type]() Transformer
	instances sync.Map
}

type transformOption struct {
	action    string
	tpe       reflect.Type
	mediaType string
}

func (v *transformers) implements(typ reflect.Type, itype reflect.Type) bool {
	return typ.Implements(itype) || reflect.PointerTo(typ).Implements(itype)
}

func (v *transformers) New(option transformOption) (Transformer, error) {
	get, _ := v.instances.LoadOrStore(option, sync.OnceValues(func() (Transformer, error) {
		mediaType := option.mediaType

		if mediaType == "" {
			switch option.tpe {
			case bytesType:
				mediaType = "octet"
			case stringType:
				mediaType = "plain"
			}

			if option.action == "marshal" {
				if v.implements(option.tpe, ioReadCloserType) {
					mediaType = "octet"
				} else if option.tpe.Implements(encodingTextMarshalerType) {
					mediaType = "plain"
				}
			} else if option.action == "unmarshal" {
				if v.implements(option.tpe, ioReadCloserType) {
					mediaType = "octet"
				} else if option.tpe.Implements(encodingTextUnmarshalerType) {
					mediaType = "plain"
				}
			}
		}

		if mediaType == "" {
			switch option.tpe.Kind() {
			case reflect.String:
				mediaType = "plain"
			default:
				mediaType = "json"
			}
		}

		p, ok := v.providers[mediaType]
		if !ok {
			return nil, fmt.Errorf("unknown supported %s", mediaType)
		}
		return p.Transformer()
	}))

	return get.(func() (Transformer, error))()
}
