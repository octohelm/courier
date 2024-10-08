package jsonschema

const (
	XEnumLabels   = `x-enum-labels`
	XGoType       = `x-go-type`
	XGoVendorType = `x-go-vendor-type`
	XGoStarLevel  = `x-go-star-level`
	XGoFieldName  = `x-go-field-name`
	XTagValidate  = `x-tag-validate`
)

// +gengo:deepcopy
type Metadata struct {
	Title       string `json:"title,omitzero"`
	Description string `json:"description,omitzero"`
	Default     any    `json:"default,omitzero"`
	WriteOnly   *bool  `json:"writeOnly,omitzero"`
	ReadOnly    *bool  `json:"readOnly,omitzero"`
	Examples    []any  `json:"examples,omitzero"`
	Deprecated  *bool  `json:"deprecated,omitzero"`

	Ext
}

func (v *Metadata) GetMetadata() *Metadata {
	return v
}

func ExtOf(ext map[string]any) Ext {
	return Ext{
		Extensions: ext,
	}
}

type Ext struct {
	Extensions map[string]any `json:",inline"`
}

func (in *Ext) DeepCopy() *Ext {
	if in == nil {
		return nil
	}
	out := new(Ext)
	in.DeepCopyInto(out)
	return out
}

func (in *Ext) DeepCopyInto(out *Ext) {
	if i := in.Extensions; i != nil {
		o := make(map[string]any, len(i))
		for key, val := range out.Extensions {
			o[key] = val
		}
		for key, val := range i {
			o[key] = val
		}
		out.Extensions = o
	}
}

func (v Ext) Merge(m Ext) Ext {
	ext := Ext{}

	for k := range v.Extensions {
		ext.AddExtension(k, v.Extensions[k])
	}

	for k := range m.Extensions {
		ext.AddExtension(k, v.Extensions[k])
	}

	return ext
}

func (v *Ext) AddExtension(key string, value any) {
	if value == nil {
		return
	}
	if v.Extensions == nil {
		v.Extensions = make(map[string]any)
	}
	v.Extensions[key] = value
}

func (v *Ext) GetExtension(key string) (any, bool) {
	if v.Extensions == nil {
		return nil, false
	}

	e, ok := v.Extensions[key]
	return e, ok
}
