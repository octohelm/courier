package v1

import "github.com/octohelm/courier/pkg/taggedunion"

// +gengo:taggedunion
// +gengo:taggedunion:underlying=Kinded
type Union struct {
	taggedunion.TaggedUnion `discriminator:"kind"`

	String *StringKinded `mapping:"String"`
	Bool   *BoolKinded   `mapping:"Bool"`
}

type Kinded interface {
	GetKind() string
}

type StringKinded struct {
	Kind string `json:"kind" validate:"@string{String}"`

	Value string `json:"value"`
}

func (StringKinded) GetKind() string {
	return "String"
}

type BoolKinded struct {
	Kind string `json:"kind" validate:"@string{Bool}"`

	Value bool `json:"value"`
}

func (BoolKinded) GetKind() string {
	return "Bool"
}
