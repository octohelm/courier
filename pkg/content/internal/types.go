package internal

import (
	"encoding"
	"io"
	"reflect"
)

var (
	stringType                  = reflect.TypeFor[string]()
	bytesType                   = reflect.TypeFor[[]byte]()
	encodingTextMarshalerType   = reflect.TypeFor[encoding.TextMarshaler]()
	encodingTextUnmarshalerType = reflect.TypeFor[encoding.TextUnmarshaler]()
	ioReadCloserType            = reflect.TypeFor[io.ReadCloser]()
)
