package org

import (
	"context"
	"net/http"
	"net/url"

	"github.com/octohelm/courier/pkg/courierhttp"
)

// 拉取组织列表
type ListOrgOld struct {
	courierhttp.MethodGet `path:"/org"`
}

func (r *ListOrgOld) Output(ctx context.Context) (any, error) {
	return courierhttp.Redirect(http.StatusFound, &url.URL{Path: "/orgs"}), nil
}
