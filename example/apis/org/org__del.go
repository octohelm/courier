package org

import (
	"context"

	"github.com/octohelm/courier/pkg/courierhttp"
)

// 删除组织
type DeleteOrg struct {
	courierhttp.MethodDelete `path:"/orgs/:orgName"`
	OrgName                  string `name:"orgName" in:"path" `
}

func (c *DeleteOrg) Output(ctx context.Context) (any, error) {
	return nil, nil
}
