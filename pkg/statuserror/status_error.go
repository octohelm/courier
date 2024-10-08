package statuserror

import "github.com/go-json-experiment/json/jsontext"

type WithStatusCode interface {
	StatusCode() int
}

type WithErrKey interface {
	ErrKey() string
}

type WithLocation interface {
	Location() string
}

type WithJSONPointer interface {
	JSONPointer() jsontext.Pointer
}
