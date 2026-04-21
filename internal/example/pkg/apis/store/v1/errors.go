package v1

import (
	"fmt"

	"github.com/octohelm/courier/pkg/statuserror"
)

// ErrBlobNotFound 表示 blob 不存在。
type ErrBlobNotFound struct {
	statuserror.NotFound

	// 对应仓库命名空间
	Namespace Namespace
	// 对应内容摘要
	Digest Digest
}

func (e ErrBlobNotFound) Error() string {
	return fmt.Sprintf("%s@%s: blob 不存在", e.Namespace, e.Digest)
}

// ErrManifestNotFound 表示 manifest 不存在。
type ErrManifestNotFound struct {
	statuserror.NotFound

	// 对应仓库命名空间
	Namespace Namespace
	// 对应 manifest 摘要
	Digest Digest
}

func (e ErrManifestNotFound) Error() string {
	return fmt.Sprintf("%s@%s: manifest 不存在", e.Namespace, e.Digest)
}

// ErrManifestInvalid 表示 manifest 内容非法。
type ErrManifestInvalid struct {
	statuserror.BadRequest

	// 错误原因
	Reason string
}

func (e ErrManifestInvalid) Error() string {
	if e.Reason == "" {
		return "manifest 非法"
	}
	return fmt.Sprintf("manifest 非法: %s", e.Reason)
}
