package org

import (
	"regexp"
	"time"

	"github.com/octohelm/courier/pkg/validator"
	"github.com/octohelm/courier/pkg/validator/validators"
	"github.com/octohelm/x/ptr"
)

type Detail struct {
	Info

	CreatedAt *time.Time `json:"createdAt,omitempty"`
}

type Name string

func (Name) StructTagValidate() string {
	return "@org-name"
}

func init() {
	validator.Register(validator.NewFormatValidatorProvider("org-name", func(format string) validator.Validator {
		return &validators.StringValidator{
			Format:        format,
			MaxLength:     ptr.Ptr[uint64](5),
			Pattern:       regexp.MustCompile(`^[a-z0-9]([-a-z0-9]*[a-z0-9])?$`),
			PatternErrMsg: "只能包含小写字母，数字和 -，且必须以小写字母或数字开头",
		}
	}))
}
