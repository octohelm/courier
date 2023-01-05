package extractors

import (
	"context"
	"sync"

	"github.com/octohelm/courier/pkg/openapi/jsonschema"
)

type OpenAPISchemaTypeGetter interface {
	OpenAPISchemaType() []string
}

type OpenAPISchemaFormatGetter interface {
	OpenAPISchemaFormat() string
}

type contextSchemaRegister struct {
}

func ContextWithSchemaRegister(ctx context.Context, sr SchemaRegister) context.Context {
	return context.WithValue(ctx, contextSchemaRegister{}, sr)
}

func SchemaRegisterFromContext(ctx context.Context) SchemaRegister {
	if v, ok := ctx.Value(contextSchemaRegister{}).(SchemaRegister); ok {
		return v
	}
	return &defaultSchemaRegister{}
}

type defaultSchemaRegister struct {
	m sync.Map
}

func (d *defaultSchemaRegister) Record(typeRef string) bool {
	_, ok := d.m.Load(typeRef)
	defer d.m.Store(typeRef, true)
	return ok
}

func (d *defaultSchemaRegister) RegisterSchema(ref string, s *jsonschema.Schema) {
	return
}

func (d *defaultSchemaRegister) RefString(ref string) string {
	return ref
}

type SchemaRegister interface {
	RegisterSchema(ref string, s *jsonschema.Schema)
	RefString(ref string) string
	Record(typeRef string) bool
}
