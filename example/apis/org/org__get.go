package org

import (
	"context"
	"github.com/octohelm/courier/example/apis/org/operator"
	"github.com/octohelm/courier/pkg/courier"
	"net/http"
	"time"

	"github.com/pkg/errors"

	"github.com/octohelm/courier/pkg/courierhttp"
	"github.com/octohelm/courier/pkg/statuserror"
)

func (GetOrg) MiddleOperators() courier.MiddleOperators {
	return courier.MiddleOperators{
		&operator.GroupOrgs{},
	}
}

// 查询组织信息
type GetOrg struct {
	courierhttp.MethodGet `path:"/:orgName"`
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
