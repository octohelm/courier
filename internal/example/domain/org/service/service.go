package service

import (
	"context"

	metav1 "github.com/octohelm/courier/internal/example/pkg/apis/meta/v1"
	orgv1 "github.com/octohelm/courier/internal/example/pkg/apis/org/v1"
)

// Service 定义组织域标准服务接口。
// +gengo:injectable:provider
type Service interface {
	// Create 创建组织。
	Create(ctx context.Context, req *orgv1.OrgForCreateRequest) (*orgv1.Org, error)
	// Update 更新组织。
	Update(ctx context.Context, orgID orgv1.OrgID, req *orgv1.OrgForUpdateRequest) (*orgv1.Org, error)
	// Delete 删除组织。
	Delete(ctx context.Context, orgID orgv1.OrgID) error
	// Get 获取组织详情。
	Get(ctx context.Context, orgID orgv1.OrgID) (*orgv1.Org, error)
	// List 查询组织列表。
	List(ctx context.Context, req *orgv1.OrgForListRequest, pager *metav1.Pager) (*orgv1.OrgList, error)
}
