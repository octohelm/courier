package openapi

import "github.com/octohelm/courier/pkg/openapi/jsonschema"

type WithContentOrSchema struct {
	Schema *Schema `json:"schema,omitempty"`
	WithContent
}

func (o *WithContentOrSchema) SetSchema(s *Schema) {
	o.Content = nil
	o.Schema = s
}

func (o *WithContentOrSchema) AddContent(contentType string, mt *MediaType) {
	o.Schema = nil
	o.WithContent.AddContent(contentType, mt)
}

type WithContent struct {
	Content map[string]*MediaType `json:"content,omitempty"`
}

func (o *WithContent) AddContent(contentType string, mt *MediaType) {
	if mt == nil {
		return
	}
	if o.Content == nil {
		o.Content = make(map[string]*MediaType)
	}
	o.Content[contentType] = mt
}

func NewMediaTypeWithSchema(s *Schema) *MediaType {
	mt := &MediaType{}
	mt.Schema = s
	return mt
}

type MediaType struct {
	MediaTypeObject
	SpecExtensions
}

func (i MediaType) MarshalJSON() ([]byte, error) {
	return jsonschema.FlattenMarshalJSON(i.MediaTypeObject, i.SpecExtensions)
}

func (i *MediaType) UnmarshalJSON(data []byte) error {
	return jsonschema.FlattenUnmarshalJSON(data, &i.MediaTypeObject, &i.SpecExtensions)
}

type MediaTypeObject struct {
	Schema  *Schema `json:"schema,omitempty"`
	Example any     `json:"example,omitempty"`
	WithExamples
	WithEncoding
}

type WithEncoding struct {
	Encoding map[string]*Encoding `json:"encoding,omitempty"`
}

func (o *WithEncoding) AddEncoding(name string, e *Encoding) {
	if e == nil {
		return
	}
	if o.Encoding == nil {
		o.Encoding = make(map[string]*Encoding)
	}
	o.Encoding[name] = e
}

func NewEncoding() *Encoding {
	return &Encoding{}
}

type Encoding struct {
	EncodingObject
	SpecExtensions
}

func (i Encoding) MarshalJSON() ([]byte, error) {
	return jsonschema.FlattenMarshalJSON(i.EncodingObject, i.SpecExtensions)
}

func (i *Encoding) UnmarshalJSON(data []byte) error {
	return jsonschema.FlattenUnmarshalJSON(data, &i.EncodingObject, &i.SpecExtensions)
}

type EncodingObject struct {
	ContentType string `json:"contentType,omitempty"`
	WithHeaders
	Style         ParameterStyle `json:"style,omitempty"`
	Explode       bool           `json:"explode,omitempty"`
	AllowReserved bool           `json:"allowReserved,omitempty"`
}
