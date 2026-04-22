package extractors

import (
	"sync"

	contextx "github.com/octohelm/x/context"

	"github.com/octohelm/courier/pkg/openapi/jsonschema"
)

var SchemaRegisterContext = contextx.New[SchemaRegister](contextx.WithDefaultsFunc(func() SchemaRegister {
	return &defaultSchemaRegister{}
}))

type defaultSchemaRegister struct{ m sync.Map }

func (d *defaultSchemaRegister) Record(typeRef string) bool {
	_, ok := d.m.Load(typeRef)
	defer d.m.Store(typeRef, true)
	return ok
}

func (d *defaultSchemaRegister) RegisterSchema(ref string, s jsonschema.Schema) {
	return
}

func (d *defaultSchemaRegister) RefString(ref string) string {
	return ref
}

type SchemaRegister interface {
	RegisterSchema(ref string, s jsonschema.Schema)
	RefString(ref string) string
	Record(typeRef string) bool
}
