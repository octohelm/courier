package openapi

// https://spec.openapis.org/oas/latest.html#pathItemObject
// no need $ref, server, parameters
type PathItemObject struct {
	Summary     string `json:"summary,omitempty"`
	Description string `json:"description,omitempty"`

	Operations map[string]*OperationObject `json:",inline"`
}

type CallbacksObject struct {
	Callbacks map[string]*PathItemObject `json:"callbacks,omitempty"`
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
