package blob

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"github.com/octohelm/courier/pkg/courierhttp"
)

type UploadBlob struct {
	courierhttp.MethodPost `path:"/blobs"`
	Blob                   io.ReadCloser `in:"body" mime:"octet-stream"`
}

func (req *UploadBlob) Output(ctx context.Context) (any, error) {
	defer req.Blob.Close()
	b := bytes.NewBuffer(nil)
	_, _ = io.Copy(b, req.Blob)
	fmt.Println(b)
	return nil, nil
}
