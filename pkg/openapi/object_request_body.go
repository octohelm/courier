package openapi

import "github.com/octohelm/courier/pkg/openapi/jsonschema"

type RequestBodyObject struct {
	Description string `json:"description,omitzero"`
	Required    bool   `json:"required,omitzero"`

	ContentObject

	jsonschema.Ext
}
