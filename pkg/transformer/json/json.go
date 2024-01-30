package json

import (
	"context"
	"encoding/json"
	"io"
	"net/textproto"
	"reflect"
	"strconv"

	transformer "github.com/octohelm/courier/pkg/transformer/core"
	"github.com/octohelm/courier/pkg/transformer/internal"
	validatorerrors "github.com/octohelm/courier/pkg/validator"
	typesutil "github.com/octohelm/x/types"
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

	return json.NewEncoder(w).Encode(v)
}

func (*jsonTransformer) DecodeFrom(ctx context.Context, r io.ReadCloser, v any, headers ...textproto.MIMEHeader) error {
	defer r.Close()

	if rv, ok := v.(reflect.Value); ok {
		if rv.Kind() != reflect.Ptr && rv.CanAddr() {
			rv = rv.Addr()
		}
		v = rv.Interface()
	}

	dec := json.NewDecoder(r)
	if err := dec.Decode(v); err != nil {
		return wrapLocationDecoderError(dec, err)
	}
	return nil
}

func wrapLocationDecoderError(dec *json.Decoder, err error) error {
	switch e := err.(type) {
	case *validatorerrors.ErrorSet:
		return e
	case *json.UnmarshalTypeError:
		r := reflect.ValueOf(dec).Elem()
		errSet := validatorerrors.NewErrorSet()
		errSet.AddErr(e, location(r.Field(1 /* .buf */).Bytes(), int(e.Offset)))
		return errSet.Err()
	case *json.SyntaxError:
		return e
	default:
		r := reflect.ValueOf(dec).Elem()
		offset := r.Field(2 /* .d */).Field(1 /* .off */).Int()
		if offset > 0 {
			errSet := validatorerrors.NewErrorSet()
			errSet.AddErr(e, location(r.Field(1 /* .buf */).Bytes(), int(offset-1)))
			return errSet.Err()
		}
		return e
	}
}

func location(data []byte, offset int) string {
	i := 0
	arrayPaths := map[string]bool{}
	arrayIdxSet := map[string]int{}
	pathWalker := internal.NewPathWalker()

	markObjectKey := func() {
		jsonKey, l := nextString(data[i:])
		i += l

		if i < int(offset) && len(jsonKey) > 0 {
			key, _ := strconv.Unquote(string(jsonKey))
			pathWalker.Enter(key)
		}
	}

	markArrayIdx := func(path string) {
		if arrayPaths[path] {
			arrayIdxSet[path]++
		} else {
			arrayPaths[path] = true
		}
		pathWalker.Enter(arrayIdxSet[path])
	}

	for i < offset {
		i += nextToken(data[i:])
		char := data[i]

		switch char {
		case '"':
			_, l := nextString(data[i:])
			i += l
		case '[', '{':
			i++

			if char == '[' {
				markArrayIdx(pathWalker.String())
			} else {
				markObjectKey()
			}
		case '}', ']', ',':
			i++
			pathWalker.Exit()

			if char == ',' {
				path := pathWalker.String()

				if _, ok := arrayPaths[path]; ok {
					markArrayIdx(path)
				} else {
					markObjectKey()
				}
			}
		default:
			i++
		}
	}

	return pathWalker.String()
}

func nextToken(data []byte) int {
	for i, c := range data {
		switch c {
		case ' ', '\n', '\r', '\t':
			continue
		default:
			return i
		}
	}
	return -1
}

func nextString(data []byte) (finalData []byte, l int) {
	quoteStartAt := -1
	for i, c := range data {
		switch c {
		case '"':
			if i > 0 && string(data[i-1]) == "\\" {
				continue
			}
			if quoteStartAt >= 0 {
				return data[quoteStartAt : i+1], i + 1
			} else {
				quoteStartAt = i
			}
		default:
			continue
		}
	}
	return nil, 0
}
