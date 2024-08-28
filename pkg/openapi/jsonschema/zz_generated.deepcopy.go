/*
Package jsonschema GENERATED BY gengo:deepcopy
DON'T EDIT THIS FILE
*/
package jsonschema

func (in *Metadata) DeepCopy() *Metadata {
	if in == nil {
		return nil
	}
	out := new(Metadata)
	in.DeepCopyInto(out)
	return out
}

func (in *Metadata) DeepCopyInto(out *Metadata) {
	out.Title = in.Title
	out.Description = in.Description
	out.Default = in.Default
	out.WriteOnly = in.WriteOnly
	out.ReadOnly = in.ReadOnly
	if in.Examples != nil {
		i, o := &in.Examples, &out.Examples
		*o = make([]any, len(*i))
		copy(*o, *i)
	}
	out.Deprecated = in.Deprecated
	in.Ext.DeepCopyInto(&out.Ext)

}
