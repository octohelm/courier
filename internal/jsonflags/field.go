package jsonflags

import (
	"encoding"
	"fmt"
	"reflect"

	"github.com/go-json-experiment/json"
	"github.com/go-json-experiment/json/jsontext"
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

	StringItem bool
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

func UnmarshalFromString(typ reflect.Type) bool {
	if Implements(typ, textUnmarshalerType) || typ == bytesType {
		return true
	}
	switch typ.Kind() {
	case reflect.String:
		return true
	case reflect.Ptr:
		return UnmarshalFromString(typ.Elem())
	default:
		return false
	}
}

func ParseFieldOptions(sf reflect.StructField) (FieldOptions, bool, error) {
	if value, ok := sf.Tag.Lookup("in"); ok && value == "body" {
		if _, ok := sf.Tag.Lookup("name"); !ok {
			sf.Tag += reflect.StructTag(fmt.Sprintf(` name:%q`, value))
		}
	}

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
		Format:     options.format,
		String:     options.string,
	}

	if !v.String {
		v.String = UnmarshalFromString(sf.Type)
	}

	if !v.String {
		if sf.Type.Kind() == reflect.Slice || sf.Type.Kind() == reflect.Slice {
			v.StringItem = UnmarshalFromString(sf.Type.Elem())
		}
	}

	return v, ignore, nil
}
