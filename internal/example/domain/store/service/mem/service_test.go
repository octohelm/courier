package mem

import (
	"bytes"
	"context"
	"io"
	"testing"

	storev1 "github.com/octohelm/courier/internal/example/pkg/apis/store/v1"
	. "github.com/octohelm/x/testing/v2"
)

func TestService(t *testing.T) {
	svc := New()
	ctx := context.Background()
	ns := storev1.Namespace("demo/project")

	t.Run("上传读取并删除 blob", func(t *testing.T) {
		desc, err := svc.UploadBlob(ctx, ns, io.NopCloser(bytes.NewBufferString("hello")))
		Then(t, "上传 blob 后返回描述对象",
			Expect(err, Equal[error](nil)),
			Expect(desc.Size, Equal(int64(5))),
		)

		body, err := svc.GetBlob(ctx, ns, desc.Digest)
		Then(t, "读取 blob 成功",
			Expect(err, Equal[error](nil)),
		)
		defer body.Close()

		raw, err := io.ReadAll(body)
		Then(t, "读取内容符合预期",
			Expect(err, Equal[error](nil)),
			Expect(string(raw), Equal("hello")),
		)

		deleted, err := svc.DeleteBlob(ctx, ns, desc.Digest)
		Then(t, "删除 blob 返回原始描述",
			Expect(err, Equal[error](nil)),
			Expect(deleted.Digest, Equal(desc.Digest)),
		)

		_, err = svc.GetBlob(ctx, ns, desc.Digest)
		Then(t, "删除后再次读取会返回不存在错误",
			ExpectDo(func() error { return err }, ErrorAsType[*storev1.ErrBlobNotFound]()),
		)
	})

	t.Run("写入读取并删除 manifest", func(t *testing.T) {
		config, err := svc.UploadBlob(ctx, ns, io.NopCloser(bytes.NewBufferString("config")))
		Then(t, "上传 config blob 成功",
			Expect(err, Equal[error](nil)),
		)
		asset, err := svc.UploadBlob(ctx, ns, io.NopCloser(bytes.NewBufferString("asset")))
		Then(t, "上传 asset blob 成功",
			Expect(err, Equal[error](nil)),
		)

		manifest := &storev1.Manifest{
			Config: *config,
			Assets: []storev1.Descriptor{*asset},
		}

		desc, err := svc.PutManifest(ctx, ns, "sha256:manifest", manifest)
		Then(t, "写入 manifest 成功",
			Expect(err, Equal[error](nil)),
			Expect(desc.Digest, Equal(storev1.Digest("sha256:manifest"))),
		)

		got, err := svc.GetManifest(ctx, ns, "sha256:manifest")
		Then(t, "读取 manifest 成功",
			Expect(err, Equal[error](nil)),
			Expect(got.Config.Digest, Equal(config.Digest)),
			Expect(len(got.Assets), Equal(1)),
		)

		err = svc.DeleteManifest(ctx, ns, "sha256:manifest")
		Then(t, "删除 manifest 成功",
			Expect(err, Equal[error](nil)),
		)

		_, err = svc.GetManifest(ctx, ns, "sha256:manifest")
		Then(t, "删除后再次读取 manifest 会返回不存在错误",
			ExpectDo(func() error { return err }, ErrorAsType[*storev1.ErrManifestNotFound]()),
		)
	})

	t.Run("缺失 blob 时 manifest 校验失败", func(t *testing.T) {
		_, err := svc.PutManifest(ctx, ns, "sha256:broken", &storev1.Manifest{
			Config: storev1.Descriptor{
				Digest: "sha256:not-exists",
			},
		})
		Then(t, "引用缺失 blob 时写入 manifest 会失败",
			ExpectDo(func() error { return err }, ErrorAsType[*storev1.ErrBlobNotFound]()),
		)
	})
}
