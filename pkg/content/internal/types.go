package internal

import (
	"encoding"
	"github.com/go-json-experiment/json"
	"io"
	"reflect"
)

var (
	stringType = reflect.TypeFor[string]()
	bytesType  = reflect.TypeFor[[]byte]()

	ioReadCloserType = reflect.TypeFor[io.ReadCloser]()

	encodingTextMarshalerType   = reflect.TypeFor[encoding.TextMarshaler]()
	encodingTextUnmarshalerType = reflect.TypeFor[encoding.TextUnmarshaler]()

	jsonUnmarshalerV1Type = reflect.TypeFor[json.UnmarshalerV1]()
	jsonUnmarshalerV2Type = reflect.TypeFor[json.UnmarshalerV2]()
)
