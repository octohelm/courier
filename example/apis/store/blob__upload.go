package store

import (
	"bytes"
	"context"
	"io"

	"github.com/octohelm/courier/pkg/courierhttp"
)

// 上传 blob
type UploadStoreBlob struct {
	courierhttp.MethodPost `path:"/{scope...}/blobs/uploads"`
	Scope                  string        `name:"scope" in:"path"`
	Blob                   io.ReadCloser `in:"body" mime:"octet-stream"`
}

func (req *UploadStoreBlob) Output(ctx context.Context) (any, error) {
	defer req.Blob.Close()
	b := bytes.NewBuffer(nil)
	_, _ = io.Copy(b, req.Blob)
	return nil, nil
}
