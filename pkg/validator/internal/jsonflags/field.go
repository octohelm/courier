package jsonflags

import (
	"encoding"
	"github.com/go-json-experiment/json"
	"github.com/go-json-experiment/json/jsontext"
	"reflect"
)

type Casing int

const (
	NoCase     Casing = 1
	StrictCase Casing = 2
)

type FieldOptions struct {
	Name       string
	QuotedName string
	HasName    bool
	Casing     Casing
	Inline     bool
	Unknown    bool
	Omitzero   bool
	Omitempty  bool
	String     bool
	Format     string
}

var (
	bytesType       = reflect.TypeFor[[]byte]()
	emptyStructType = reflect.TypeFor[struct{}]()
)

var (
	jsontextValueType     = reflect.TypeFor[jsontext.Value]()
	textUnmarshalerType   = reflect.TypeFor[encoding.TextUnmarshaler]()
	jsonUnmarshalerV1Type = reflect.TypeFor[json.UnmarshalerV1]()
	jsonUnmarshalerV2Type = reflect.TypeFor[json.UnmarshalerV2]()
)

func ParseFieldOptions(sf reflect.StructField) (FieldOptions, bool, error) {
	options, ignore, err := parseFieldOptions(sf)
	if err != nil {
		return FieldOptions{}, ignore, err
	}

	v := FieldOptions{
		Name:       options.name,
		QuotedName: options.quotedName,
		HasName:    options.hasName,
		Casing:     Casing(options.casing),
		Inline:     options.inline,
		Unknown:    options.unknown,
		Omitzero:   options.omitzero,
		Omitempty:  options.omitempty,
		String:     options.string,
		Format:     options.format,
	}

	if !v.String {
		if Implements(sf.Type, textUnmarshalerType) || sf.Type == bytesType {
			v.String = true
		} else {
			switch sf.Type.Kind() {
			case reflect.String:
				v.String = true
			case reflect.Ptr:
				if sf.Type.Elem().Kind() == reflect.String {
					v.String = true
				}
			}
		}
	}

	return v, ignore, nil
}
