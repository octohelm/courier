package v1

import (
	"regexp"
	"time"

	metav1 "github.com/octohelm/courier/internal/example/pkg/apis/meta/v1"
	"github.com/octohelm/courier/pkg/validator"
	"github.com/octohelm/courier/pkg/validator/validators"
)

type OrgList = metav1.List[Org]

// Org 表示组织对象。
type Org struct {
	// 组织 ID
	ID OrgID `json:"id"`
	// 创建时间
	CreatedAt *time.Time `json:"createdAt,omitempty"`
	// 更新时间
	UpdatedAt *time.Time `json:"updatedAt,omitempty"`

	// 组织规格
	Spec OrgSpec `json:"spec"`
}

// +gengo:uintstr
type OrgID uint64

// OrgSpec 表示组织规格。
type OrgSpec struct {
	// 组织名称
	Name OrgName `json:"name"`
	// 组织类型
	Type OrgType `json:"type,omitzero"`
}

// OrgName 表示组织名称。
type OrgName string

func (OrgName) StructTagValidate() string {
	return "@org-name"
}

func init() {
	// 自定义 strfmt
	validator.Register(validator.NewFormatValidatorProvider("org-name", func(format string) validator.Validator {
		return &validators.StringValidator{
			Format:        format,
			MaxLength:     new(uint64(5)),
			Pattern:       regexp.MustCompile(`^[a-z0-9]([-a-z0-9]*[a-z0-9])?$`),
			PatternErrMsg: "只能包含小写字母，数字和 -，且必须以小写字母或数字开头",
		}
	}))
}
