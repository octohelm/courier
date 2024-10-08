package openapi

import "github.com/octohelm/courier/pkg/openapi/jsonschema"

type EncodingObject struct {
	ContentType string `json:"contentType"`

	HeadersObject

	Style         ParameterStyle `json:"style,omitzero"`
	Explode       bool           `json:"explode,omitzero"`
	AllowReserved bool           `json:"allowReserved,omitzero"`

	jsonschema.Ext
}
