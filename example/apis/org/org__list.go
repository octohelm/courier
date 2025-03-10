package org

import (
	"context"
	"net/http"

	"github.com/octohelm/courier/example/pkg/filter"
	"github.com/octohelm/courier/pkg/courierhttp"
)

// 拉取组织列表
type ListOrg struct {
	courierhttp.MethodGet `path:"/orgs"`

	ID *filter.Filter[ID] `json:"org~id,omitzero" in:"query"`
}

func (r *ListOrg) Output(ctx context.Context) (any, error) {
	return courierhttp.Wrap(
		&DataList[Info]{},
		courierhttp.WithStatusCode(http.StatusOK),
		courierhttp.WithMetadata("X-Custom", "X"),
	), nil
}

type ID int64

type DataList[T any] struct {
	Data  []T   `json:"data"`
	Total int   `json:"total"`
	Extra []any `json:"extra"`
}
