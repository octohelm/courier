package openapi

import "github.com/octohelm/courier/pkg/openapi/jsonschema"

type OperationGetter interface {
	OpenAPIOperation(ref func(t string) jsonschema.Refer) *Operation
}
