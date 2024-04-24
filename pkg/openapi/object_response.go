package openapi

import (
	"fmt"

	"github.com/octohelm/courier/pkg/openapi/jsonschema"
)

type ResponseObject struct {
	Description string `json:"description"`

	HeadersObject
	ContentObject

	jsonschema.Ext
}

type ResponsesObject struct {
	Responses map[string]*ResponseObject `json:"responses"`
}

func (o *ResponsesObject) SetDefaultResponse(r *ResponseObject) {
	if r == nil {
		return
	}
	if o.Responses == nil {
		o.Responses = make(map[string]*ResponseObject)
	}
	o.Responses["default"] = r
}

func (o *ResponsesObject) AddResponse(statusCode int, r *ResponseObject) {
	if r == nil {
		return
	}
	if o.Responses == nil {
		o.Responses = make(map[string]*ResponseObject)
	}
	o.Responses[fmt.Sprintf("%d", statusCode)] = r
}
