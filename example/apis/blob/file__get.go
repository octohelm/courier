package blob

import (
	"context"
	"github.com/octohelm/courier/pkg/courierhttp"
)

type GetFile struct {
	courierhttp.MethodPost `path:"/blobs/{path...}"`

	Path string `name:"path" in:"path"`
}

func (req *GetFile) Output(ctx context.Context) (any, error) {
	return req.Path, nil
}
