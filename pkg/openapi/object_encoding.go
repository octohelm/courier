package openapi

import "github.com/octohelm/courier/pkg/openapi/jsonschema"

type EncodingObject struct {
	ContentType string `json:"contentType"`

	HeadersObject

	Style         ParameterStyle `json:"style,omitempty"`
	Explode       bool           `json:"explode,omitempty"`
	AllowReserved bool           `json:"allowReserved,omitempty"`

	jsonschema.Ext
}
