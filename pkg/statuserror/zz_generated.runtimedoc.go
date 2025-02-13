/*
Package statuserror GENERATED BY gengo:runtimedoc
DON'T EDIT THIS FILE
*/
package statuserror

import _ "embed"

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

func (v *Descriptor) RuntimeDoc(names ...string) ([]string, bool) {
	if len(names) > 0 {
		switch names[0] {
		case "Code":
			return []string{
				"错误编码",
			}, true
		case "Message":
			return []string{
				"错误信息",
			}, true
		case "Description":
			return []string{
				"错误详情",
			}, true
		case "Location":
			return []string{
				"错误参数位置 query, header, path, body 等",
			}, true
		case "Pointer":
			return []string{
				"错误参数 json pointer",
			}, true
		case "Source":
			return []string{
				"引起错误的源",
			}, true
		case "Errors":
			return []string{
				"错误链",
			}, true
		case "Extra":
			return []string{}, true
		case "Status":
			return []string{}, true

		}

		return nil, false
	}
	return []string{}, true
}

func (v *ErrorResponse) RuntimeDoc(names ...string) ([]string, bool) {
	if len(names) > 0 {
		switch names[0] {
		case "Code":
			return []string{
				"错误状态码",
			}, true
		case "Msg":
			return []string{
				"错误信息",
			}, true
		case "Errors":
			return []string{
				"错误详情",
			}, true

		}

		return nil, false
	}
	return []string{}, true
}

func (*IntOrString) RuntimeDoc(names ...string) ([]string, bool) {
	return []string{}, true
}
