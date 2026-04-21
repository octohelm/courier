package v1

import (
	"io"

	"github.com/octohelm/courier/pkg/courierhttp"

	storev1 "github.com/octohelm/courier/internal/example/pkg/apis/store/v1"
)

// UploadBlob 表示上传 blob。
type UploadBlob struct {
	courierhttp.MethodPost `path:"/v1/{namespace...}/blobs/uploads"`
	// 仓库命名空间
	Namespace storev1.Namespace `name:"namespace" in:"path"`
	// 二进制内容
	Body io.ReadCloser `in:"body" mime:"application/octet-stream"`
}

func (UploadBlob) ResponseData() *storev1.Descriptor {
	return new(storev1.Descriptor)
}

// DeleteBlob 表示删除 blob。
type DeleteBlob struct {
	courierhttp.MethodDelete `path:"/v1/{namespace...}/blobs/{digest}"`

	// 仓库命名空间
	Namespace storev1.Namespace `name:"namespace" in:"path"`
	// 内容摘要
	Digest storev1.Digest `name:"digest" in:"path"`
}

func (DeleteBlob) ResponseData() *storev1.Descriptor {
	return new(storev1.Descriptor)
}

func (DeleteBlob) ResponseErrors() []error {
	return []error{
		&storev1.ErrBlobNotFound{},
	}
}

// GetBlob 表示拉取 blob。
type GetBlob struct {
	courierhttp.MethodGet `path:"/v1/{namespace...}/blobs/{digest}"`

	// 仓库命名空间
	Namespace storev1.Namespace `name:"namespace" in:"path"`
	// 内容摘要
	Digest storev1.Digest `name:"digest" in:"path"`
}

func (GetBlob) ResponseData() io.ReadCloser {
	return nil
}

func (GetBlob) ResponseErrors() []error {
	return []error{
		&storev1.ErrBlobNotFound{},
	}
}
