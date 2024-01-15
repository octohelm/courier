package openapi

import "github.com/octohelm/courier/pkg/openapi/jsonschema"

type RequestBodyObject struct {
	Description string `json:"description,omitempty"`
	Required    bool   `json:"required,omitempty"`

	ContentObject

	jsonschema.Ext
}
