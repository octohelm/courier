package org

import (
	"context"
	"fmt"

	"github.com/octohelm/courier/example/pkg/domain/org"
	"github.com/octohelm/courier/pkg/courierhttp"
	"github.com/octohelm/courier/pkg/validator"
)

// 创建组织
type CreateOrg struct {
	courierhttp.MethodPost `path:"/orgs"`

	Info `in:"body"`
}

func (c *CreateOrg) Output(ctx context.Context) (interface{}, error) {
	req, _ := courierhttp.RequestFromContext(ctx)
	fmt.Println(req.ContentLength)
	fmt.Println(c.Info)

	return nil, nil
}

// 组织详情
type Info struct {
	// 组织名称
	Name Name `json:"name"`
	// 组织类型
	Type org.Type `json:"type,omitzero"`
}

type info2 Info

func (info Info) MarshalJSON() ([]byte, error) {
	return validator.Marshal(info2(info))
}
