/*
Package store GENERATED BY gengo:runtimedoc
DON'T EDIT THIS FILE
*/
package store

func (v *GetStoreBlob) RuntimeDoc(names ...string) ([]string, bool) {
	if len(names) > 0 {
		switch names[0] {
		case "Scope":
			return []string{}, true
		case "Digest":
			return []string{}, true

		}

		return nil, false
	}
	return []string{
		"获取 blob",
	}, true
}

func (v *UploadStoreBlob) RuntimeDoc(names ...string) ([]string, bool) {
	if len(names) > 0 {
		switch names[0] {
		case "Scope":
			return []string{}, true
		case "Blob":
			return []string{}, true

		}

		return nil, false
	}
	return []string{
		"上传 blob",
	}, true
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
