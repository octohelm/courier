/*
Package internal GENERATED BY gengo:runtimedoc 
DON'T EDIT THIS FILE
*/
package internal

// nolint:deadcode,unused
func runtimeDoc(v any, names ...string) ([]string, bool) {
	if c, ok := v.(interface {
		RuntimeDoc(names ...string) ([]string, bool)
	}); ok {
		return c.RuntimeDoc(names...)
	}
	return nil, false
}

func (v ParamValue) RuntimeDoc(names ...string) ([]string, bool) {
	if len(names) > 0 {
		switch names[0] {

		}

		return nil, false
	}
	return []string{}, true
}

func (v Request) RuntimeDoc(names ...string) ([]string, bool) {
	if len(names) > 0 {
		switch names[0] {
		case "ParamValue":
			return []string{}, true

		}
		if doc, ok := runtimeDoc(v.ParamValue, names...); ok {
			return doc, ok
		}

		return nil, false
	}
	return []string{}, true
}
