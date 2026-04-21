package v1

import (
	"fmt"

	"github.com/octohelm/courier/pkg/statuserror"
)

// ErrOrgNotFound 表示组织不存在。
type ErrOrgNotFound struct {
	statuserror.NotFound

	// 查询时使用的组织 ID
	OrgID *OrgID
	// 查询时使用的组织名称
	OrgName OrgName
}

func (e ErrOrgNotFound) Error() string {
	if e.OrgID != nil {
		return fmt.Sprintf("%d: 组织不存在", *e.OrgID)
	}
	return fmt.Sprintf("%s: 组织不存在", e.OrgName)
}

// ErrOrgNameConflict 表示组织名称冲突。
type ErrOrgNameConflict struct {
	statuserror.Conflict

	// 已存在的组织名称
	OrgName OrgName
}

func (e ErrOrgNameConflict) Error() string {
	return fmt.Sprintf("%s: 组织名称已存在", e.OrgName)
}
