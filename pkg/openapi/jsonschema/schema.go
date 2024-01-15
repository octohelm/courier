package jsonschema

import (
	"bytes"
	"github.com/go-json-experiment/json"
	"github.com/go-json-experiment/json/jsontext"
	"io"
)

type OpenAPISchemaGetter interface {
	OpenAPISchema() Schema
}

type OpenAPISchemaTypeGetter interface {
	OpenAPISchemaType() []string
}

type OpenAPISchemaFormatGetter interface {
	OpenAPISchemaFormat() string
}

// interface of k8s pkgs
type CanSwaggerDoc interface {
	SwaggerDoc() map[string]string
}

func IsType[T Schema](s any) bool {
	if _, ok := s.(T); ok {
		return true
	}
	return false
}

type Payload struct {
	Schema
}

func Unmarshal(data []byte, v any) error {
	if err := json.UnmarshalDecode(jsontext.NewDecoder(bytes.NewReader(data)), v, json.WithUnmarshalers(schemaUnmarshalers)); err != nil {
		return err
	}
	return nil
}

func (p Payload) MarshalJSON() ([]byte, error) {
	return json.Marshal(p.Schema)
}

func (p *Payload) UnmarshalJSON(data []byte) (err error) {
	var schema Schema
	if err := json.UnmarshalDecode(jsontext.NewDecoder(bytes.NewReader(data)), &schema, json.WithUnmarshalers(schemaUnmarshalers)); err != nil {
		return err
	}
	*p = Payload{
		Schema: schema,
	}
	return nil
}

type Schema interface {
	GetCore() *Core
	GetMetadata() *Metadata
	PrintTo(w io.Writer)
}
