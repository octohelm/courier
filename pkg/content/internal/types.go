package internal

import (
	"encoding"
	"io"
	"reflect"

	"encoding/json/v2"
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
