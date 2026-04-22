package v1

import (
	storev1 "github.com/octohelm/courier/internal/example/pkg/apis/store/v1"
	"github.com/octohelm/courier/pkg/courier"
	"github.com/octohelm/courier/pkg/courierhttp"
)

// PutManifest 表示写入 manifest。
type PutManifest struct {
	courierhttp.MethodPut `path:"/v1/{namespace...}/manifests/{digest}"`

	// 仓库命名空间
	Namespace storev1.Namespace `name:"namespace" in:"path"`
	// manifest 摘要
	Digest storev1.Digest `name:"digest" in:"path"`

	// manifest 内容
	Body *storev1.Manifest `in:"body"`
}

func (PutManifest) ResponseData() *storev1.Descriptor {
	return new(storev1.Descriptor)
}

func (PutManifest) ResponseErrors() []error {
	return []error{
		&storev1.ErrBlobNotFound{},
		&storev1.ErrManifestInvalid{},
	}
}

// GetManifest 表示拉取 manifest。
type GetManifest struct {
	courierhttp.MethodGet `path:"/v1/{namespace...}/manifests/{digest}"`

	// 仓库命名空间
	Namespace storev1.Namespace `name:"namespace" in:"path"`
	// manifest 摘要
	Digest storev1.Digest `name:"digest" in:"path"`
}

func (GetManifest) ResponseData() *storev1.Manifest {
	return new(storev1.Manifest)
}

func (GetManifest) ResponseErrors() []error {
	return []error{
		&storev1.ErrManifestNotFound{},
	}
}

// DeleteManifest 表示删除 manifest。
type DeleteManifest struct {
	courierhttp.MethodDelete `path:"/v1/{namespace...}/manifests/{digest}"`

	// 仓库命名空间
	Namespace storev1.Namespace `name:"namespace" in:"path"`
	// manifest 摘要
	Digest storev1.Digest `name:"digest" in:"path"`
}

func (DeleteManifest) ResponseData() *courier.NoContent {
	return new(courier.NoContent)
}

func (DeleteManifest) ResponseErrors() []error {
	return []error{
		&storev1.ErrManifestNotFound{},
	}
}
