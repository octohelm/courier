/*
Package jsonschema GENERATED BY gengo:runtimedoc 
DON'T EDIT THIS FILE
*/
package jsonschema

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

func (AnchorString) RuntimeDoc(names ...string) ([]string, bool) {
	return []string{
		"openapi:strfmt anchor",
	}, true
}

func (v AnyType) RuntimeDoc(names ...string) ([]string, bool) {
	if len(names) > 0 {
		switch names[0] {

		}
		if doc, ok := runtimeDoc(v.Core, "", names...); ok {
			return doc, ok
		}
		if doc, ok := runtimeDoc(v.Metadata, "", names...); ok {
			return doc, ok
		}

		return nil, false
	}
	return []string{}, true
}

func (v ArrayType) RuntimeDoc(names ...string) ([]string, bool) {
	if len(names) > 0 {
		switch names[0] {
		case "Type":
			return []string{}, true
		case "Items":
			return []string{}, true
		case "MaxItems":
			return []string{
				"validate",
			}, true
		case "MinItems":
			return []string{}, true
		case "UniqueItems":
			return []string{}, true

		}
		if doc, ok := runtimeDoc(v.Core, "", names...); ok {
			return doc, ok
		}
		if doc, ok := runtimeDoc(v.Metadata, "", names...); ok {
			return doc, ok
		}

		return nil, false
	}
	return []string{}, true
}

func (v BooleanType) RuntimeDoc(names ...string) ([]string, bool) {
	if len(names) > 0 {
		switch names[0] {
		case "Type":
			return []string{}, true

		}
		if doc, ok := runtimeDoc(v.Core, "", names...); ok {
			return doc, ok
		}
		if doc, ok := runtimeDoc(v.Metadata, "", names...); ok {
			return doc, ok
		}

		return nil, false
	}
	return []string{}, true
}

func (v Core) RuntimeDoc(names ...string) ([]string, bool) {
	if len(names) > 0 {
		switch names[0] {
		case "Schema":
			return []string{
				"default https://json-schema.org/draft/2020-12/schema",
			}, true
		case "ID":
			return []string{}, true
		case "Comment":
			return []string{}, true
		case "Vocabulary":
			return []string{}, true
		case "Anchor":
			return []string{
				"https://json-schema.org/understanding-json-schema/structuring#anchor",
			}, true
		case "DynamicAnchor":
			return []string{
				"for generics type",
				"https://json-schema.org/blog/posts/dynamicref-and-generics",
			}, true
		case "Defs":
			return []string{}, true

		}

		return nil, false
	}
	return []string{}, true
}

func (v Discriminator) RuntimeDoc(names ...string) ([]string, bool) {
	if len(names) > 0 {
		switch names[0] {
		case "PropertyName":
			return []string{}, true
		case "Mapping":
			return []string{}, true

		}

		return nil, false
	}
	return []string{}, true
}

func (v EnumType) RuntimeDoc(names ...string) ([]string, bool) {
	if len(names) > 0 {
		switch names[0] {
		case "Enum":
			return []string{}, true

		}
		if doc, ok := runtimeDoc(v.Core, "", names...); ok {
			return doc, ok
		}
		if doc, ok := runtimeDoc(v.Metadata, "", names...); ok {
			return doc, ok
		}

		return nil, false
	}
	return []string{}, true
}

func (v Ext) RuntimeDoc(names ...string) ([]string, bool) {
	if len(names) > 0 {
		switch names[0] {
		case "Extensions":
			return []string{}, true

		}

		return nil, false
	}
	return []string{}, true
}

func (v IntersectionType) RuntimeDoc(names ...string) ([]string, bool) {
	if len(names) > 0 {
		switch names[0] {
		case "AllOf":
			return []string{}, true

		}
		if doc, ok := runtimeDoc(v.Core, "", names...); ok {
			return doc, ok
		}
		if doc, ok := runtimeDoc(v.Metadata, "", names...); ok {
			return doc, ok
		}

		return nil, false
	}
	return []string{}, true
}

func (v Metadata) RuntimeDoc(names ...string) ([]string, bool) {
	if len(names) > 0 {
		switch names[0] {
		case "Title":
			return []string{}, true
		case "Description":
			return []string{}, true
		case "Default":
			return []string{}, true
		case "WriteOnly":
			return []string{}, true
		case "ReadOnly":
			return []string{}, true
		case "Examples":
			return []string{}, true
		case "Deprecated":
			return []string{}, true

		}
		if doc, ok := runtimeDoc(v.Ext, "", names...); ok {
			return doc, ok
		}

		return nil, false
	}
	return []string{}, true
}

func (v NullType) RuntimeDoc(names ...string) ([]string, bool) {
	if len(names) > 0 {
		switch names[0] {
		case "Type":
			return []string{}, true

		}
		if doc, ok := runtimeDoc(v.Core, "", names...); ok {
			return doc, ok
		}
		if doc, ok := runtimeDoc(v.Metadata, "", names...); ok {
			return doc, ok
		}

		return nil, false
	}
	return []string{}, true
}

func (v NumberType) RuntimeDoc(names ...string) ([]string, bool) {
	if len(names) > 0 {
		switch names[0] {
		case "Type":
			return []string{}, true
		case "MultipleOf":
			return []string{
				"validate",
			}, true
		case "Maximum":
			return []string{}, true
		case "ExclusiveMaximum":
			return []string{}, true
		case "Minimum":
			return []string{}, true
		case "ExclusiveMinimum":
			return []string{}, true

		}
		if doc, ok := runtimeDoc(v.Core, "", names...); ok {
			return doc, ok
		}
		if doc, ok := runtimeDoc(v.Metadata, "", names...); ok {
			return doc, ok
		}

		return nil, false
	}
	return []string{}, true
}

func (v ObjectType) RuntimeDoc(names ...string) ([]string, bool) {
	if len(names) > 0 {
		switch names[0] {
		case "Type":
			return []string{}, true
		case "Properties":
			return []string{}, true
		case "PropertyNames":
			return []string{}, true
		case "AdditionalProperties":
			return []string{}, true
		case "Required":
			return []string{
				"validate",
			}, true
		case "MaxProperties":
			return []string{}, true
		case "MinProperties":
			return []string{}, true

		}
		if doc, ok := runtimeDoc(v.Core, "", names...); ok {
			return doc, ok
		}
		if doc, ok := runtimeDoc(v.Metadata, "", names...); ok {
			return doc, ok
		}

		return nil, false
	}
	return []string{}, true
}

func (v Payload) RuntimeDoc(names ...string) ([]string, bool) {
	if len(names) > 0 {
		switch names[0] {

		}
		if doc, ok := runtimeDoc(v.Schema, "", names...); ok {
			return doc, ok
		}

		return nil, false
	}
	return []string{}, true
}

func (v RefType) RuntimeDoc(names ...string) ([]string, bool) {
	if len(names) > 0 {
		switch names[0] {
		case "Ref":
			return []string{}, true
		case "DynamicRef":
			return []string{}, true

		}
		if doc, ok := runtimeDoc(v.Core, "", names...); ok {
			return doc, ok
		}
		if doc, ok := runtimeDoc(v.Metadata, "", names...); ok {
			return doc, ok
		}

		return nil, false
	}
	return []string{}, true
}

func (SchemaPrintOption) RuntimeDoc(names ...string) ([]string, bool) {
	return []string{}, true
}

func (v StringType) RuntimeDoc(names ...string) ([]string, bool) {
	if len(names) > 0 {
		switch names[0] {
		case "Type":
			return []string{}, true
		case "Format":
			return []string{}, true
		case "MaxLength":
			return []string{
				"validate",
			}, true
		case "MinLength":
			return []string{}, true
		case "Pattern":
			return []string{}, true

		}
		if doc, ok := runtimeDoc(v.Core, "", names...); ok {
			return doc, ok
		}
		if doc, ok := runtimeDoc(v.Metadata, "", names...); ok {
			return doc, ok
		}

		return nil, false
	}
	return []string{}, true
}

func (v URIReferenceString) RuntimeDoc(names ...string) ([]string, bool) {
	if len(names) > 0 {
		switch names[0] {
		case "Scheme":
			return []string{}, true
		case "Opaque":
			return []string{}, true
		case "User":
			return []string{
				"encoded opaque data",
			}, true
		case "Host":
			return []string{
				"username and password information",
			}, true
		case "Path":
			return []string{
				"host or host:port (see Hostname and Port methods)",
			}, true
		case "RawPath":
			return []string{
				"path (relative paths may omit leading slash)",
			}, true
		case "OmitHost":
			return []string{
				"encoded path hint (see EscapedPath method)",
			}, true
		case "ForceQuery":
			return []string{
				"do not emit empty host (authority)",
			}, true
		case "RawQuery":
			return []string{
				"append a query ('?') even if RawQuery is empty",
			}, true
		case "Fragment":
			return []string{
				"encoded query values, without '?'",
			}, true
		case "RawFragment":
			return []string{
				"fragment for references, without '#'",
			}, true

		}

		return nil, false
	}
	return []string{
		"openapi:strfmt uri-reference",
	}, true
}

func (v URIString) RuntimeDoc(names ...string) ([]string, bool) {
	if len(names) > 0 {
		switch names[0] {
		case "Scheme":
			return []string{}, true
		case "Opaque":
			return []string{}, true
		case "User":
			return []string{
				"encoded opaque data",
			}, true
		case "Host":
			return []string{
				"username and password information",
			}, true
		case "Path":
			return []string{
				"host or host:port (see Hostname and Port methods)",
			}, true
		case "RawPath":
			return []string{
				"path (relative paths may omit leading slash)",
			}, true
		case "OmitHost":
			return []string{
				"encoded path hint (see EscapedPath method)",
			}, true
		case "ForceQuery":
			return []string{
				"do not emit empty host (authority)",
			}, true
		case "RawQuery":
			return []string{
				"append a query ('?') even if RawQuery is empty",
			}, true
		case "Fragment":
			return []string{
				"encoded query values, without '?'",
			}, true
		case "RawFragment":
			return []string{
				"fragment for references, without '#'",
			}, true

		}

		return nil, false
	}
	return []string{
		"openapi:strfmt uri",
	}, true
}

func (v UnionType) RuntimeDoc(names ...string) ([]string, bool) {
	if len(names) > 0 {
		switch names[0] {
		case "OneOf":
			return []string{}, true
		case "Discriminator":
			return []string{}, true

		}
		if doc, ok := runtimeDoc(v.Core, "", names...); ok {
			return doc, ok
		}
		if doc, ok := runtimeDoc(v.Metadata, "", names...); ok {
			return doc, ok
		}

		return nil, false
	}
	return []string{}, true
}
