package openapi

import "github.com/octohelm/courier/pkg/openapi/jsonschema"

type ContentObject struct {
	Content map[string]*MediaTypeObject `json:"content,omitempty"`
}

func (o *ContentObject) AddContent(contentType string, mt *MediaTypeObject) {
	if mt == nil {
		return
	}
	if o.Content == nil {
		o.Content = make(map[string]*MediaTypeObject)
	}
	o.Content[contentType] = mt
}

type MediaTypeObject struct {
	Schema   jsonschema.Schema          `json:"schema,omitempty"`
	Encoding map[string]*EncodingObject `json:"encoding,omitempty"`

	jsonschema.Ext
}

func (o *MediaTypeObject) AddEncoding(name string, e *EncodingObject) {
	if e == nil {
		return
	}
	if o.Encoding == nil {
		o.Encoding = make(map[string]*EncodingObject)
	}
	o.Encoding[name] = e
}
