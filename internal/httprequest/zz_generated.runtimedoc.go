/*
Package httprequest GENERATED BY gengo:runtimedoc 
DON'T EDIT THIS FILE
*/
package httprequest

// nolint:deadcode,unused
func runtimeDoc(v any, names ...string) ([]string, bool) {
	if c, ok := v.(interface {
		RuntimeDoc(names ...string) ([]string, bool)
	}); ok {
		return c.RuntimeDoc(names...)
	}
	return nil, false
}

func (Params) RuntimeDoc(names ...string) ([]string, bool) {
	return []string{}, true
}