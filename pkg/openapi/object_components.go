package openapi

import (
	"github.com/octohelm/courier/pkg/openapi/jsonschema"
	"github.com/octohelm/courier/pkg/ptr"
	"net/url"
	"strings"
)

// https://spec.openapis.org/oas/latest.html#components-object
// FIXME now only support schemas
type ComponentsObject struct {
	Schemas map[string]jsonschema.Schema `json:"schemas,omitempty"`
}

func (o *ComponentsObject) AddSchema(id string, s jsonschema.Schema) {
	if s == nil {
		return
	}
	if o.Schemas == nil {
		o.Schemas = make(map[string]jsonschema.Schema)
	}
	o.Schemas[id] = s
}

func (o *ComponentsObject) RefSchema(id string) jsonschema.Schema {
	if o.Schemas == nil || o.Schemas[id] == nil {
		return nil
	}
	return &jsonschema.RefType{
		Ref: ptr.Ptr(jsonschema.URIReferenceString(url.URL{
			Fragment: strings.Join([]string{"#", "components", "schemas", id}, "/"),
		})),
	}
}
