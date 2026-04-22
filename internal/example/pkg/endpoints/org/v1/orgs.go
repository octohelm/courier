package v1

import (
	metav1 "github.com/octohelm/courier/internal/example/pkg/apis/meta/v1"
	orgv1 "github.com/octohelm/courier/internal/example/pkg/apis/org/v1"
	"github.com/octohelm/courier/pkg/courier"
	"github.com/octohelm/courier/pkg/courierhttp"
)

// 创建组织
type CreateOrg struct {
	courierhttp.MethodPost `path:"/v1/orgs"`

	// 创建参数
	Body orgv1.OrgForCreateRequest `in:"body"`
}

func (CreateOrg) ResponseData() *orgv1.Org {
	return new(orgv1.Org)
}

func (CreateOrg) ResponseErrors() []error {
	return []error{
		&orgv1.ErrOrgNameConflict{},
	}
}

// 更新组织
type UpdateOrg struct {
	courierhttp.MethodPatch `path:"/v1/orgs/{orgID}"`

	// 组织 ID
	OrgID orgv1.OrgID `name:"orgID" in:"path" `
	// 更新参数
	Body orgv1.OrgForUpdateRequest `in:"body"`
}

func (UpdateOrg) ResponseData() *orgv1.Org {
	return new(orgv1.Org)
}

func (UpdateOrg) ResponseErrors() []error {
	return []error{
		&orgv1.ErrOrgNotFound{},
		&orgv1.ErrOrgNameConflict{},
	}
}

// 删除组织
type DeleteOrg struct {
	courierhttp.MethodDelete `path:"/v1/orgs/{orgID}"`

	// 组织 ID
	OrgID orgv1.OrgID `name:"orgID" in:"path" `
}

func (DeleteOrg) ResponseData() *courier.NoContent {
	return new(courier.NoContent)
}

func (DeleteOrg) ResponseErrors() []error {
	return []error{
		&orgv1.ErrOrgNotFound{},
	}
}

// 查询组织信息
type GetOrg struct {
	courierhttp.MethodGet `path:"/v1/orgs/{orgID}"`

	// 组织 ID
	OrgID orgv1.OrgID `name:"orgID" in:"path" `
}

func (GetOrg) ResponseData() *orgv1.Org {
	return new(orgv1.Org)
}

func (GetOrg) ResponseErrors() []error {
	return []error{
		&orgv1.ErrOrgNotFound{},
	}
}

// 拉取组织列表
type ListOrg struct {
	courierhttp.MethodGet `path:"/v1/orgs"`

	// 按组织 ID 过滤
	OrgID *orgv1.OrgID `name:"org~id,omitzero" in:"query"`
	// 按组织名称过滤
	OrgName *orgv1.OrgName `name:"org~name,omitzero" in:"query"`
	// 按组织类型过滤
	OrgType *orgv1.OrgType `name:"org~type,omitzero" in:"query"`

	metav1.Pager
}

func (ListOrg) ResponseData() *orgv1.OrgList {
	return new(orgv1.OrgList)
}
