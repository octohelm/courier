/*
Package example GENERATED BY gengo:runtimedoc 
DON'T EDIT THIS FILE
*/
package example

// nolint:deadcode,unused
func runtimeDoc(v any, names ...string) ([]string, bool) {
	if c, ok := v.(interface {
		RuntimeDoc(names ...string) ([]string, bool)
	}); ok {
		return c.RuntimeDoc(names...)
	}
	return nil, false
}

func (v Client) RuntimeDoc(names ...string) ([]string, bool) {
	if len(names) > 0 {
		switch names[0] {
		case "Endpoint":
			return []string{}, true
		case "HttpTransports":
			return []string{}, true

		}

		return nil, false
	}
	return []string{}, true
}

func (v Cookie) RuntimeDoc(names ...string) ([]string, bool) {
	if len(names) > 0 {
		switch names[0] {
		case "Token":
			return []string{}, true

		}

		return nil, false
	}
	return []string{}, true
}

func (v CreateOrg) RuntimeDoc(names ...string) ([]string, bool) {
	if len(names) > 0 {
		switch names[0] {
		case "OrgInfo":
			return []string{}, true

		}
		if doc, ok := runtimeDoc(v.OrgInfo, names...); ok {
			return doc, ok
		}

		return nil, false
	}
	return []string{}, true
}

func (v DeleteOrg) RuntimeDoc(names ...string) ([]string, bool) {
	if len(names) > 0 {
		switch names[0] {
		case "OrgName":
			return []string{}, true

		}

		return nil, false
	}
	return []string{}, true
}

func (v GetOrg) RuntimeDoc(names ...string) ([]string, bool) {
	if len(names) > 0 {
		switch names[0] {
		case "OrgName":
			return []string{}, true

		}

		return nil, false
	}
	return []string{}, true
}

func (v OrgDetail) RuntimeDoc(names ...string) ([]string, bool) {
	if len(names) > 0 {
		switch names[0] {
		case "CreatedAt":
			return []string{}, true
		case "Name":
			return []string{
				"组织名称",
			}, true
		case "Type":
			return []string{
				"组织类型",
			}, true

		}

		return nil, false
	}
	return []string{}, true
}

func (v GetStoreBlob) RuntimeDoc(names ...string) ([]string, bool) {
	if len(names) > 0 {
		switch names[0] {
		case "Scope":
			return []string{}, true
		case "Digest":
			return []string{}, true

		}

		return nil, false
	}
	return []string{}, true
}

func (v ListOrg) RuntimeDoc(names ...string) ([]string, bool) {
	if len(names) > 0 {
		switch names[0] {

		}

		return nil, false
	}
	return []string{}, true
}

func (v ListOrgOld) RuntimeDoc(names ...string) ([]string, bool) {
	if len(names) > 0 {
		switch names[0] {

		}

		return nil, false
	}
	return []string{}, true
}

func (v OrgDataList) RuntimeDoc(names ...string) ([]string, bool) {
	if len(names) > 0 {
		switch names[0] {
		case "Data":
			return []string{}, true
		case "Extra":
			return []string{}, true
		case "Total":
			return []string{}, true

		}

		return nil, false
	}
	return []string{}, true
}

func (v OrgInfo) RuntimeDoc(names ...string) ([]string, bool) {
	if len(names) > 0 {
		switch names[0] {
		case "Name":
			return []string{
				"组织名称",
			}, true
		case "Type":
			return []string{
				"组织类型",
			}, true

		}

		return nil, false
	}
	return []string{}, true
}

func (OrgType) RuntimeDoc(names ...string) ([]string, bool) {
	return []string{}, true
}
func (Time) RuntimeDoc(names ...string) ([]string, bool) {
	return []string{}, true
}
func (v UploadBlob) RuntimeDoc(names ...string) ([]string, bool) {
	if len(names) > 0 {
		switch names[0] {
		case "ReadCloser":
			return []string{}, true

		}
		if doc, ok := runtimeDoc(v.ReadCloser, names...); ok {
			return doc, ok
		}

		return nil, false
	}
	return []string{}, true
}

func (v UploadStoreBlob) RuntimeDoc(names ...string) ([]string, bool) {
	if len(names) > 0 {
		switch names[0] {
		case "Scope":
			return []string{}, true
		case "ReadCloser":
			return []string{}, true

		}
		if doc, ok := runtimeDoc(v.ReadCloser, names...); ok {
			return doc, ok
		}

		return nil, false
	}
	return []string{}, true
}
