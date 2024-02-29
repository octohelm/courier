package operator

import (
	"context"
	"github.com/octohelm/courier/pkg/courierhttp"
)

type GroupOrgs struct {
	courierhttp.Method `path:"/orgs"`
}

func (GroupOrgs) Output(ctx context.Context) (any, error) {
	return nil, nil
}
