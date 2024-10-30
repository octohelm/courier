package internal

import (
	"context"
	"fmt"
	"io"
	"reflect"
	"strings"
	"sync"
)

type TransformerProvider interface {
	Names() []string
	Transformer() (Transformer, error)
}

type Transformer interface {
	MediaType() string

	ContentReader
	ContentProvider
}

type ContentReader interface {
	ReadAs(ctx context.Context, r io.ReadCloser, v any) error
}

type Content interface {
	GetContentType() string
	GetContentLength() int64

	io.ReadCloser
}

type ContentProvider interface {
	Prepare(ctx context.Context, src any) (Content, error)
}

type ContentLengthSetter interface {
	SetContentLength(l int64)
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

		if strings.HasSuffix(mediaType, "+json") {
			mediaType = "json"
		}

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
				} else if reflect.PointerTo(option.tpe).Implements(encodingTextUnmarshalerType) {
					mediaType = "plain"
				} else if reflect.PointerTo(option.tpe).Implements(jsonUnmarshalerV1Type) || reflect.PointerTo(option.tpe).Implements(jsonUnmarshalerV2Type) {
					mediaType = "json"
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
			return nil, fmt.Errorf("unknown media type %s to %s", mediaType, option.tpe.String())
		}
		return p.Transformer()
	}))

	return get.(func() (Transformer, error))()
}
