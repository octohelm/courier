/*
Package errors GENERATED BY gengo:runtimedoc
DON'T EDIT THIS FILE
*/
package errors

func (v *ErrInvalidType) RuntimeDoc(names ...string) ([]string, bool) {
	if len(names) > 0 {
		switch names[0] {
		case "Type":
			return []string{}, true
		case "Value":
			return []string{}, true

		}

		return nil, false
	}
	return []string{}, true
}

func (v *ErrMultipleOf) RuntimeDoc(names ...string) ([]string, bool) {
	if len(names) > 0 {
		switch names[0] {
		case "Topic":
			return []string{}, true
		case "Current":
			return []string{}, true
		case "MultipleOf":
			return []string{}, true

		}

		return nil, false
	}
	return []string{}, true
}

func (v *ErrNotMatch) RuntimeDoc(names ...string) ([]string, bool) {
	if len(names) > 0 {
		switch names[0] {
		case "Topic":
			return []string{}, true
		case "Current":
			return []string{}, true
		case "Pattern":
			return []string{}, true

		}

		return nil, false
	}
	return []string{}, true
}

func (v *NotInEnumError) RuntimeDoc(names ...string) ([]string, bool) {
	if len(names) > 0 {
		switch names[0] {
		case "Topic":
			return []string{}, true
		case "Current":
			return []string{}, true
		case "Enums":
			return []string{}, true

		}

		return nil, false
	}
	return []string{}, true
}

func (v *OutOfRangeError) RuntimeDoc(names ...string) ([]string, bool) {
	if len(names) > 0 {
		switch names[0] {
		case "Topic":
			return []string{}, true
		case "Current":
			return []string{}, true
		case "Minimum":
			return []string{}, true
		case "Maximum":
			return []string{}, true
		case "ExclusiveMaximum":
			return []string{}, true
		case "ExclusiveMinimum":
			return []string{}, true

		}

		return nil, false
	}
	return []string{}, true
}

// nolint:deadcode,unused
func runtimeDoc(v any, prefix string, names ...string) ([]string, bool) {
	if c, ok := v.(interface {
		RuntimeDoc(names ...string) ([]string, bool)
	}); ok {
		doc, ok := c.RuntimeDoc(names...)
		if ok {
			if prefix != "" && len(doc) > 0 {
				doc[0] = prefix + doc[0]
				return doc, true
			}

			return doc, true
		}
	}
	return nil, false
}
