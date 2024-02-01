package store

import (
	"context"
	"fmt"
	"github.com/octohelm/courier/pkg/courierhttp"
)

// 获取 blob
type GetStoreBlob struct {
	courierhttp.MethodGet `path:"/{scope...}/blobs/{digest}"`
	Scope                 string `name:"scope" in:"path"`
	Digest                string `name:"digest" in:"path"`
}

func (req *GetStoreBlob) Output(ctx context.Context) (any, error) {
	return fmt.Sprintf("%s@%s", req.Scope, req.Digest), nil
}
