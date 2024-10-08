package jsonschema

type Core struct {
	// Ref    		 string    `json:"$ref,omitzero"`
	// DynamicRef    string    `json:"$dynamicRef,omitzero"`

	// default https://json-schema.org/draft/2020-12/schema
	Schema *URIString `json:"$schema,omitzero"`

	ID *URIReferenceString `json:"$id,omitzero"`

	Comment string `json:"$comment,omitzero"`

	Vocabulary map[URIString]bool `json:"$vocabulary,omitzero"`

	// https://json-schema.org/understanding-json-schema/structuring#anchor
	Anchor string `json:"$anchor,omitzero"`

	// for generics type
	// https://json-schema.org/blog/posts/dynamicref-and-generics
	DynamicAnchor string `json:"$dynamicAnchor,omitzero"`

	Defs map[string]Schema `json:"$defs,omitzero"`
}

func (core *Core) GetCore() *Core {
	return core
}
