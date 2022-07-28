package org

import (
	"context"

	"github.com/octohelm/courier/pkg/courierhttp"
)

// 创建组织
type CreateOrg struct {
	courierhttp.MethodPost `path:"/orgs"`
	Info                   `in:"body"`
}

func (c *CreateOrg) Output(ctx context.Context) (interface{}, error) {
	return nil, nil
}

// 组织详情
type Info struct {
	// 组织名称
	Name string `json:"name" validate:"@string[0,5]"`
	// 组织类型
	Type Type `json:"type,omitempty"`
}
