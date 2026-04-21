package service

import (
	"context"
	"io"

	storev1 "github.com/octohelm/courier/internal/example/pkg/apis/store/v1"
)

// Service 定义制品仓库域标准服务接口。
// +gengo:injectable:provider
type Service interface {
	// UploadBlob 上传 blob。
	UploadBlob(ctx context.Context, namespace storev1.Namespace, body io.ReadCloser) (*storev1.Descriptor, error)
	// GetBlob 拉取 blob。
	GetBlob(ctx context.Context, namespace storev1.Namespace, digest storev1.Digest) (io.ReadCloser, error)
	// DeleteBlob 删除 blob。
	DeleteBlob(ctx context.Context, namespace storev1.Namespace, digest storev1.Digest) (*storev1.Descriptor, error)
	// PutManifest 写入 manifest。
	PutManifest(ctx context.Context, namespace storev1.Namespace, digest storev1.Digest, manifest *storev1.Manifest) (*storev1.Descriptor, error)
	// GetManifest 拉取 manifest。
	GetManifest(ctx context.Context, namespace storev1.Namespace, digest storev1.Digest) (*storev1.Manifest, error)
	// DeleteManifest 删除 manifest。
	DeleteManifest(ctx context.Context, namespace storev1.Namespace, digest storev1.Digest) error
}
