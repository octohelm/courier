/*
Package pathpattern GENERATED BY gengo:runtimedoc 
DON'T EDIT THIS FILE
*/
package pathpattern

// nolint:deadcode,unused
func runtimeDoc(v any, names ...string) ([]string, bool) {
	if c, ok := v.(interface {
		RuntimeDoc(names ...string) ([]string, bool)
	}); ok {
		return c.RuntimeDoc(names...)
	}
	return nil, false
}

func (v Route) RuntimeDoc(names ...string) ([]string, bool) {
	if len(names) > 0 {
		switch names[0] {
		case "Method":
			return []string{}, true
		case "PathSegments":
			return []string{}, true
		case "ChildSegments":
			return []string{}, true

		}

		return nil, false
	}
	return []string{}, true
}

func (Segments) RuntimeDoc(names ...string) ([]string, bool) {
	return []string{}, true
}
func (Values) RuntimeDoc(names ...string) ([]string, bool) {
	return []string{}, true
}
