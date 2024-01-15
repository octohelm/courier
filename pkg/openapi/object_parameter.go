package openapi

import (
	"github.com/octohelm/courier/pkg/openapi/jsonschema"
	"github.com/octohelm/x/ptr"
)

func NewParameter(name string, in ParameterIn) *ParameterObject {
	return &ParameterObject{
		Name: name,
		In:   in,
	}
}

// https://spec.openapis.org/oas/latest.html#parameter-object
type ParameterObject struct {
	Name string      `json:"name"`
	In   ParameterIn `json:"in"`

	Parameter
}

type Parameter struct {
	Schema      jsonschema.Schema `json:"schema"`
	Description string            `json:"description,omitempty"`
	Required    *bool             `json:"required,omitempty"`
	Deprecated  *bool             `json:"deprecated,omitempty"`

	// https://spec.openapis.org/oas/latest.html#parameter-object
	Style   ParameterStyle `json:"style,omitempty"`
	Explode *bool          `json:"explode,omitempty"`

	jsonschema.Ext
}

type HeaderObject = Parameter

type HeadersObject struct {
	Headers map[string]*HeaderObject `json:"headers,omitempty"`
}

func (object *HeadersObject) AddHeader(name string, h *Parameter) {
	if h == nil {
		return
	}
	if object.Headers == nil {
		object.Headers = make(map[string]*Parameter)
	}

	object.Headers[name] = h
}

func (o *ParameterObject) SetDefaultStyle() {
	switch o.In {
	case InPath, InHeader:
		o.Style = ParameterStyleSimple
	case InQuery, InCookie:
		o.Style = ParameterStyleForm
	}

	switch o.Style {
	case ParameterStyleForm:
		o.Explode = ptr.Bool(true)
	}
}

type ParameterIn string

const (
	InQuery  ParameterIn = "query"
	InPath   ParameterIn = "path"
	InHeader ParameterIn = "header"
	InCookie ParameterIn = "cookie"
)

type ParameterStyle string

const (
	// https://tools.ietf.org/html/rfc6570#section-3.2.7
	ParameterStyleMatrix ParameterStyle = "matrix"
	// https://tools.ietf.org/html/rfc6570#section-3.2.5
	ParameterStyleLabel ParameterStyle = "label"
	// https://tools.ietf.org/html/rfc6570#section-3.2.8
	ParameterStyleForm ParameterStyle = "form"
	// for array, csv https://tools.ietf.org/html/rfc6570#section-3.2.2
	ParameterStyleSimple ParameterStyle = "simple"
	// for array, ssv
	ParameterStyleSpaceDelimited ParameterStyle = "spaceDelimited"
	// for array, pipes
	ParameterStylePipeDelimited ParameterStyle = "pipeDelimited"
	// for object
	ParameterStyleDeepObject ParameterStyle = "deepObject"
)
