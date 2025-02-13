package internal

import (
	"encoding"
	"io"
	"reflect"

	"github.com/go-json-experiment/json"
)

var (
	stringType = reflect.TypeFor[string]()
	bytesType  = reflect.TypeFor[[]byte]()

	ioReadCloserType = reflect.TypeFor[io.ReadCloser]()

	encodingTextMarshalerType   = reflect.TypeFor[encoding.TextMarshaler]()
	encodingTextUnmarshalerType = reflect.TypeFor[encoding.TextUnmarshaler]()

	jsonUnmarshalerType     = reflect.TypeFor[json.Unmarshaler]()
	jsonUnmarshalerFromType = reflect.TypeFor[json.UnmarshalerFrom]()
)
