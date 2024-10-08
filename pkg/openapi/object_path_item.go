package openapi

// https://spec.openapis.org/oas/latest.html#pathItemObject
// no need $ref, server, parameters
type PathItemObject struct {
	Summary     string `json:"summary,omitzero"`
	Description string `json:"description,omitzero"`

	Operations map[string]*OperationObject `json:",inline"`
}

type CallbacksObject struct {
	Callbacks map[string]*PathItemObject `json:"callbacks,omitzero"`
}

func (o *CallbacksObject) AddCallback(name string, c *PathItemObject) {
	if c == nil {
		return
	}
	if o.Callbacks == nil {
		o.Callbacks = make(map[string]*PathItemObject)
	}

	o.Callbacks[name] = c
}
