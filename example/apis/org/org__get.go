package org

import (
	"context"
	"net/http"
	"time"

	"github.com/octohelm/courier/pkg/statuserror"
	"github.com/pkg/errors"

	"github.com/octohelm/courier/pkg/courierhttp"
)

// 查询组织信息
type GetOrg struct {
	courierhttp.MethodGet `path:"/orgs/:orgName"`
	OrgName               string `name:"orgName" in:"path" `
}

func (c *GetOrg) Output(ctx context.Context) (any, error) {
	if c.OrgName == "NotFound" {
		return nil, statuserror.Wrap(errors.New("NotFound"), http.StatusNotFound, "NotFound")
	}

	return &Detail{
		Info: Info{
			Name: c.OrgName,
			Type: TYPE__GOV,
		},
	}, nil
}

type Detail struct {
	Info
	CreatedAt *time.Time `json:"createdAt,omitempty"`
}
