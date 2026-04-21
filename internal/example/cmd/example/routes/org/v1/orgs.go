package v1

import (
	"context"

	orgservice "github.com/octohelm/courier/internal/example/domain/org/service"
	orgv1 "github.com/octohelm/courier/internal/example/pkg/apis/org/v1"
	endpointorgv1 "github.com/octohelm/courier/internal/example/pkg/endpoints/org/v1"
)

// +gengo:injectable
type CreateOrg struct {
	endpointorgv1.CreateOrg

	svc orgservice.Service `inject:""`
}

func (r *CreateOrg) Output(ctx context.Context) (any, error) {
	return r.svc.Create(ctx, &r.Body)
}

// +gengo:injectable
type UpdateOrg struct {
	endpointorgv1.UpdateOrg

	svc orgservice.Service `inject:""`
}

func (r *UpdateOrg) Output(ctx context.Context) (any, error) {
	return r.svc.Update(ctx, r.OrgID, &r.Body)
}

// +gengo:injectable
type DeleteOrg struct {
	endpointorgv1.DeleteOrg

	svc orgservice.Service `inject:""`
}

func (r *DeleteOrg) Output(ctx context.Context) (any, error) {
	return nil, r.svc.Delete(ctx, r.OrgID)
}

// +gengo:injectable
type GetOrg struct {
	endpointorgv1.GetOrg

	svc orgservice.Service `inject:""`
}

func (r *GetOrg) Output(ctx context.Context) (any, error) {
	return r.svc.Get(ctx, r.OrgID)
}

// +gengo:injectable
type ListOrg struct {
	endpointorgv1.ListOrg

	svc orgservice.Service `inject:""`
}

func (r *ListOrg) Output(ctx context.Context) (any, error) {
	return r.svc.List(ctx, &orgv1.OrgForListRequest{
		OrgID:   r.OrgID,
		OrgName: r.OrgName,
		OrgType: r.OrgType,
	}, &r.Pager)
}
