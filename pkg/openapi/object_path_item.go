package openapi

// https://spec.openapis.org/oas/latest.html#pathItemObject
// no need $ref, server, parameters
type PathItemObject struct {
	Summary     string `json:"summary,omitempty"`
	Description string `json:"description,omitempty"`

	GET     *OperationObject `json:"get,omitempty"`
	PUT     *OperationObject `json:"put,omitempty"`
	POST    *OperationObject `json:"post,omitempty"`
	DELETE  *OperationObject `json:"delete,omitempty"`
	OPTIONS *OperationObject `json:"options,omitempty"`
	HEAD    *OperationObject `json:"head,omitempty"`
	PATCH   *OperationObject `json:"patch,omitempty"`
	TRACE   *OperationObject `json:"trace,omitempty"`
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
