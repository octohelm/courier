/*
Package jsonflags GENERATED BY gengo:runtimedoc 
DON'T EDIT THIS FILE
*/
package jsonflags

// nolint:deadcode,unused
func runtimeDoc(v any, names ...string) ([]string, bool) {
	if c, ok := v.(interface {
		RuntimeDoc(names ...string) ([]string, bool)
	}); ok {
		return c.RuntimeDoc(names...)
	}
	return nil, false
}

func (Casing) RuntimeDoc(names ...string) ([]string, bool) {
	return []string{}, true
}
func (v FieldOptions) RuntimeDoc(names ...string) ([]string, bool) {
	if len(names) > 0 {
		switch names[0] {
		case "Name":
			return []string{}, true
		case "QuotedName":
			return []string{}, true
		case "HasName":
			return []string{}, true
		case "Casing":
			return []string{}, true
		case "Inline":
			return []string{}, true
		case "Unknown":
			return []string{}, true
		case "Omitzero":
			return []string{}, true
		case "Omitempty":
			return []string{}, true
		case "String":
			return []string{}, true
		case "Format":
			return []string{}, true
		case "StringItem":
			return []string{}, true

		}

		return nil, false
	}
	return []string{}, true
}

func (v StructField) RuntimeDoc(names ...string) ([]string, bool) {
	if len(names) > 0 {
		switch names[0] {
		case "FieldOptions":
			return []string{}, true
		case "FieldName":
			return []string{}, true
		case "Tag":
			return []string{}, true
		case "Type":
			return []string{}, true

		}
		if doc, ok := runtimeDoc(v.FieldOptions, names...); ok {
			return doc, ok
		}

		return nil, false
	}
	return []string{}, true
}
