/*
Package openapi GENERATED BY gengo:runtimedoc 
DON'T EDIT THIS FILE
*/
package openapi

// nolint:deadcode,unused
func runtimeDoc(v any, names ...string) ([]string, bool) {
	if c, ok := v.(interface {
		RuntimeDoc(names ...string) ([]string, bool)
	}); ok {
		return c.RuntimeDoc(names...)
	}
	return nil, false
}

func (v CallbacksObject) RuntimeDoc(names ...string) ([]string, bool) {
	if len(names) > 0 {
		switch names[0] {
		case "Callbacks":
			return []string{}, true

		}

		return nil, false
	}
	return []string{}, true
}

func (v ComponentsObject) RuntimeDoc(names ...string) ([]string, bool) {
	if len(names) > 0 {
		switch names[0] {
		case "Schemas":
			return []string{}, true

		}

		return nil, false
	}
	return []string{
		"https://spec.openapis.org/oas/latest.html#components-object",
		"FIXME now only support schemas",
	}, true
}

func (v Contact) RuntimeDoc(names ...string) ([]string, bool) {
	if len(names) > 0 {
		switch names[0] {
		case "Name":
			return []string{}, true
		case "URL":
			return []string{}, true
		case "Email":
			return []string{}, true

		}

		return nil, false
	}
	return []string{}, true
}

func (v ContentObject) RuntimeDoc(names ...string) ([]string, bool) {
	if len(names) > 0 {
		switch names[0] {
		case "Content":
			return []string{}, true

		}

		return nil, false
	}
	return []string{}, true
}

func (v EncodingObject) RuntimeDoc(names ...string) ([]string, bool) {
	if len(names) > 0 {
		switch names[0] {
		case "ContentType":
			return []string{}, true
		case "HeadersObject":
			return []string{}, true
		case "Style":
			return []string{}, true
		case "Explode":
			return []string{}, true
		case "AllowReserved":
			return []string{}, true
		case "Ext":
			return []string{}, true

		}
		if doc, ok := runtimeDoc(v.HeadersObject, names...); ok {
			return doc, ok
		}
		if doc, ok := runtimeDoc(v.Ext, names...); ok {
			return doc, ok
		}

		return nil, false
	}
	return []string{}, true
}

func (v HeadersObject) RuntimeDoc(names ...string) ([]string, bool) {
	if len(names) > 0 {
		switch names[0] {
		case "Headers":
			return []string{}, true

		}

		return nil, false
	}
	return []string{}, true
}

func (v InfoObject) RuntimeDoc(names ...string) ([]string, bool) {
	if len(names) > 0 {
		switch names[0] {
		case "Title":
			return []string{}, true
		case "Description":
			return []string{}, true
		case "TermsOfService":
			return []string{}, true
		case "Contact":
			return []string{}, true
		case "License":
			return []string{}, true
		case "Version":
			return []string{}, true

		}

		return nil, false
	}
	return []string{
		"https://spec.openapis.org/oas/latest.html#infoObject",
	}, true
}

func (v License) RuntimeDoc(names ...string) ([]string, bool) {
	if len(names) > 0 {
		switch names[0] {
		case "Name":
			return []string{}, true
		case "URL":
			return []string{}, true

		}

		return nil, false
	}
	return []string{}, true
}

func (v MediaTypeObject) RuntimeDoc(names ...string) ([]string, bool) {
	if len(names) > 0 {
		switch names[0] {
		case "Schema":
			return []string{}, true
		case "Encoding":
			return []string{}, true
		case "Ext":
			return []string{}, true

		}
		if doc, ok := runtimeDoc(v.Ext, names...); ok {
			return doc, ok
		}

		return nil, false
	}
	return []string{}, true
}

func (v OpenAPI) RuntimeDoc(names ...string) ([]string, bool) {
	if len(names) > 0 {
		switch names[0] {
		case "OpenAPI":
			return []string{}, true
		case "InfoObject":
			return []string{}, true
		case "ComponentsObject":
			return []string{}, true
		case "Paths":
			return []string{}, true
		case "Ext":
			return []string{}, true

		}
		if doc, ok := runtimeDoc(v.InfoObject, names...); ok {
			return doc, ok
		}
		if doc, ok := runtimeDoc(v.ComponentsObject, names...); ok {
			return doc, ok
		}
		if doc, ok := runtimeDoc(v.Ext, names...); ok {
			return doc, ok
		}

		return nil, false
	}
	return []string{}, true
}

func (v OperationObject) RuntimeDoc(names ...string) ([]string, bool) {
	if len(names) > 0 {
		switch names[0] {
		case "Tags":
			return []string{}, true
		case "Summary":
			return []string{}, true
		case "Description":
			return []string{}, true
		case "OperationId":
			return []string{}, true
		case "Parameters":
			return []string{}, true
		case "RequestBody":
			return []string{}, true
		case "ResponsesObject":
			return []string{}, true
		case "CallbacksObject":
			return []string{}, true
		case "Deprecated":
			return []string{}, true
		case "Ext":
			return []string{}, true

		}
		if doc, ok := runtimeDoc(v.ResponsesObject, names...); ok {
			return doc, ok
		}
		if doc, ok := runtimeDoc(v.CallbacksObject, names...); ok {
			return doc, ok
		}
		if doc, ok := runtimeDoc(v.Ext, names...); ok {
			return doc, ok
		}

		return nil, false
	}
	return []string{}, true
}

func (v Parameter) RuntimeDoc(names ...string) ([]string, bool) {
	if len(names) > 0 {
		switch names[0] {
		case "Schema":
			return []string{}, true
		case "Description":
			return []string{}, true
		case "Required":
			return []string{}, true
		case "Deprecated":
			return []string{}, true
		case "Style":
			return []string{
				"https://spec.openapis.org/oas/latest.html#parameter-object",
			}, true
		case "Explode":
			return []string{}, true
		case "Ext":
			return []string{}, true

		}
		if doc, ok := runtimeDoc(v.Ext, names...); ok {
			return doc, ok
		}

		return nil, false
	}
	return []string{}, true
}

func (ParameterIn) RuntimeDoc(names ...string) ([]string, bool) {
	return []string{}, true
}
func (v ParameterObject) RuntimeDoc(names ...string) ([]string, bool) {
	if len(names) > 0 {
		switch names[0] {
		case "Name":
			return []string{}, true
		case "In":
			return []string{}, true
		case "Parameter":
			return []string{}, true

		}
		if doc, ok := runtimeDoc(v.Parameter, names...); ok {
			return doc, ok
		}

		return nil, false
	}
	return []string{
		"https://spec.openapis.org/oas/latest.html#parameter-object",
	}, true
}

func (ParameterStyle) RuntimeDoc(names ...string) ([]string, bool) {
	return []string{}, true
}
func (v PathItemObject) RuntimeDoc(names ...string) ([]string, bool) {
	if len(names) > 0 {
		switch names[0] {
		case "Summary":
			return []string{}, true
		case "Description":
			return []string{}, true
		case "Operations":
			return []string{}, true

		}

		return nil, false
	}
	return []string{
		"https://spec.openapis.org/oas/latest.html#pathItemObject",
		"no need $ref, server, parameters",
	}, true
}

func (v Payload) RuntimeDoc(names ...string) ([]string, bool) {
	if len(names) > 0 {
		switch names[0] {
		case "OpenAPI":
			return []string{}, true

		}
		if doc, ok := runtimeDoc(v.OpenAPI, names...); ok {
			return doc, ok
		}

		return nil, false
	}
	return []string{}, true
}

func (v RequestBodyObject) RuntimeDoc(names ...string) ([]string, bool) {
	if len(names) > 0 {
		switch names[0] {
		case "Description":
			return []string{}, true
		case "Required":
			return []string{}, true
		case "ContentObject":
			return []string{}, true
		case "Ext":
			return []string{}, true

		}
		if doc, ok := runtimeDoc(v.ContentObject, names...); ok {
			return doc, ok
		}
		if doc, ok := runtimeDoc(v.Ext, names...); ok {
			return doc, ok
		}

		return nil, false
	}
	return []string{}, true
}

func (v ResponseObject) RuntimeDoc(names ...string) ([]string, bool) {
	if len(names) > 0 {
		switch names[0] {
		case "Description":
			return []string{}, true
		case "HeadersObject":
			return []string{}, true
		case "ContentObject":
			return []string{}, true
		case "Ext":
			return []string{}, true

		}
		if doc, ok := runtimeDoc(v.HeadersObject, names...); ok {
			return doc, ok
		}
		if doc, ok := runtimeDoc(v.ContentObject, names...); ok {
			return doc, ok
		}
		if doc, ok := runtimeDoc(v.Ext, names...); ok {
			return doc, ok
		}

		return nil, false
	}
	return []string{}, true
}

func (v ResponsesObject) RuntimeDoc(names ...string) ([]string, bool) {
	if len(names) > 0 {
		switch names[0] {
		case "Responses":
			return []string{}, true

		}

		return nil, false
	}
	return []string{}, true
}
