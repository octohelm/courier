package jsonschema

type Core struct {
	// Ref    		 string    `json:"$ref,omitempty"`
	// DynamicRef    string    `json:"$dynamicRef,omitempty"`

	// default https://json-schema.org/draft/2020-12/schema
	Schema *URIString `json:"$schema,omitempty"`

	ID *URIReferenceString `json:"$id,omitempty"`

	Comment string `json:"$comment,omitempty"`

	Vocabulary map[URIString]bool `json:"$vocabulary,omitempty"`

	// https://json-schema.org/understanding-json-schema/structuring#anchor
	Anchor string `json:"$anchor,omitempty"`

	// for generics type
	// https://json-schema.org/blog/posts/dynamicref-and-generics
	DynamicAnchor string `json:"$dynamicAnchor,omitempty"`

	Defs map[string]Schema `json:"$defs,omitempty"`
}

func (core *Core) GetCore() *Core {
	return core
}
