package org

import (
	"context"
	"time"

	"github.com/octohelm/courier/example/pkg/domain/org"

	"github.com/octohelm/courier/example/apis/org/operator"
	"github.com/octohelm/courier/pkg/courier"
	"github.com/octohelm/courier/pkg/courierhttp"
)

func (GetOrg) MiddleOperators() courier.MiddleOperators {
	return courier.MiddleOperators{
		&operator.GroupOrgs{},
	}
}

// 查询组织信息
type GetOrg struct {
	courierhttp.MethodGet `path:"/:orgName"`

	OrgName string `name:"orgName" in:"path" `
}

func (c *GetOrg) Output(ctx context.Context) (any, error) {
	if c.OrgName == "NotFound" {
		return nil, &ErrNotFound{OrgName: c.OrgName}
	}

	return &Detail{
		Info: Info{
			Name: c.OrgName,
			Type: org.TYPE__GOV,
		},
	}, nil
}

type Detail struct {
	Info
	CreatedAt *time.Time `json:"createdAt,omitempty"`
}
