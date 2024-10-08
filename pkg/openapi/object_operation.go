package openapi

import "github.com/octohelm/courier/pkg/openapi/jsonschema"

func NewOperation(operationId string) *OperationObject {
	op := &OperationObject{}
	op.OperationId = operationId
	return op
}

type OperationObject struct {
	Tags []string `json:"tags,omitzero"`

	Summary     string `json:"summary,omitzero"`
	Description string `json:"description,omitzero"`

	OperationId string `json:"operationId"`

	Parameters  []*ParameterObject `json:"parameters,omitzero"`
	RequestBody *RequestBodyObject `json:"requestBody,omitzero"`

	ResponsesObject

	CallbacksObject

	Deprecated *bool `json:"deprecated,omitzero"`

	jsonschema.Ext
}

func (o OperationObject) WithTags(tags ...string) *OperationObject {
	o.Tags = append(o.Tags, tags...)
	return &o
}

func (o OperationObject) WithSummary(summary string) *OperationObject {
	o.Summary = summary
	return &o
}

func (o OperationObject) WithDesc(desc string) *OperationObject {
	o.Description = desc
	return &o
}

func (o *OperationObject) SetRequestBody(rb *RequestBodyObject) {
	o.RequestBody = rb
}

func (o *OperationObject) AddParameter(name string, in ParameterIn, p *Parameter) {
	if p == nil {
		return
	}
	o.Parameters = append(o.Parameters, &ParameterObject{
		Name:      name,
		In:        in,
		Parameter: *p,
	})
}
