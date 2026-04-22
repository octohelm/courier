package routes

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/octohelm/x/testing/v2"

	orgservice "github.com/octohelm/courier/internal/example/domain/org/service"
	orgmem "github.com/octohelm/courier/internal/example/domain/org/service/mem"
	storeservice "github.com/octohelm/courier/internal/example/domain/store/service"
	storemem "github.com/octohelm/courier/internal/example/domain/store/service/mem"
	orgv1 "github.com/octohelm/courier/internal/example/pkg/apis/org/v1"
	storev1 "github.com/octohelm/courier/internal/example/pkg/apis/store/v1"
	endpointorgv1 "github.com/octohelm/courier/internal/example/pkg/endpoints/org/v1"
	endpointstorev1 "github.com/octohelm/courier/internal/example/pkg/endpoints/store/v1"
	"github.com/octohelm/courier/pkg/courier"
	"github.com/octohelm/courier/pkg/courierhttp/client"
	"github.com/octohelm/courier/pkg/courierhttp/handler"
	"github.com/octohelm/courier/pkg/courierhttp/handler/httprouter"
	"github.com/octohelm/courier/pkg/statuserror"
)

func TestRoutes(t *testing.T) {
	ctx := context.Background()

	srv := newTestServer(t)
	defer srv.Close()

	orgClient := newOrgClient(srv)
	storeClient := newStoreClient(srv)

	t.Run("组织接口请求回路完整", func(t *testing.T) {
		created, err := courier.DoWith(ctx, orgClient, &endpointorgv1.CreateOrg{
			Body: orgv1.OrgForCreateRequest{
				Spec: orgv1.OrgSpec{
					Name: "alpha",
					Type: orgv1.ORG_TYPE__GOV,
				},
			},
		})
		Then(t, "创建组织接口可返回新建对象",
			Expect(err, Equal[error](nil)),
			Expect(created.ID > 0, Equal(true)),
			Expect(created.Spec.Name, Equal(orgv1.OrgName("alpha"))),
		)

		listed, err := courier.DoWith(ctx, orgClient, &endpointorgv1.ListOrg{
			OrgType: ptr(orgv1.ORG_TYPE__GOV),
		})
		Then(t, "列表接口可命中过滤条件",
			Expect(err, Equal[error](nil)),
			Expect(listed.Total, Equal(int64(1))),
			Expect(len(listed.Items), Equal(1)),
			Expect(listed.Items[0].ID, Equal(created.ID)),
		)

		got, err := courier.DoWith(ctx, orgClient, &endpointorgv1.GetOrg{
			OrgID: created.ID,
		})

		Then(t, "详情接口可读取已创建组织",
			Expect(err, Equal[error](nil)),
			Expect(got.Spec.Name, Equal(orgv1.OrgName("alpha"))),
		)

		updated, err := courier.DoWith(ctx, orgClient, &endpointorgv1.UpdateOrg{
			OrgID: created.ID,
			Body: orgv1.OrgForUpdateRequest{
				Spec: orgv1.OrgSpec{
					Name: "beta2",
					Type: orgv1.ORG_TYPE__COMPANY,
				},
			},
		})
		Then(t, "更新接口可返回最新组织内容",
			Expect(err, Equal[error](nil)),
			Expect(updated.Spec.Name, Equal(orgv1.OrgName("beta2"))),
			Expect(updated.Spec.Type, Equal(orgv1.ORG_TYPE__COMPANY)),
		)

		_, err = courier.DoWith(ctx, orgClient, &endpointorgv1.CreateOrg{
			Body: orgv1.OrgForCreateRequest{
				Spec: orgv1.OrgSpec{
					Name: "beta2",
					Type: orgv1.ORG_TYPE__GOV,
				},
			},
		})
		Then(t, "重复组织名会返回冲突错误",
			ExpectDo(func() error { return err }, ErrorAsType[*statuserror.Descriptor]()),
			ExpectMust(func() error { return expectStatusCode(err, http.StatusConflict) }),
		)

		_, err = courier.DoWith(ctx, orgClient, &endpointorgv1.DeleteOrg{
			OrgID: created.ID,
		})
		Then(t, "删除接口可成功执行",
			Expect(err, Equal[error](nil)),
		)

		_, err = courier.DoWith(ctx, orgClient, &endpointorgv1.GetOrg{
			OrgID: created.ID,
		})
		Then(t, "删除后再次读取会返回不存在错误",
			ExpectDo(func() error { return err }, ErrorAsType[*statuserror.Descriptor]()),
			ExpectMust(func() error { return expectStatusCode(err, http.StatusNotFound) }),
		)
	})

	t.Run("仓库接口请求回路完整", func(t *testing.T) {
		ns := storev1.Namespace("demo/project")

		config, err := courier.DoWith(ctx, storeClient, &endpointstorev1.UploadBlob{
			Namespace: ns,
			Body:      ioReadCloser("config"),
		})
		Then(t, "上传 config blob 成功",
			Expect(err, Equal[error](nil)),
			Expect(config.Size, Equal(int64(6))),
		)

		asset, err := courier.DoWith(ctx, storeClient, &endpointstorev1.UploadBlob{
			Namespace: ns,
			Body:      ioReadCloser("asset"),
		})
		Then(t, "上传 asset blob 成功",
			Expect(err, Equal[error](nil)),
			Expect(asset.Size, Equal(int64(5))),
		)

		r, err := courier.DoWith(ctx, storeClient, &endpointstorev1.GetBlob{
			Namespace: ns,
			Digest:    config.Digest,
		})
		raw := MustValue(t, func() ([]byte, error) {
			defer r.Close()
			return io.ReadAll(r)
		})
		Then(t, "读取 blob 接口可返回原始内容",
			Expect(err, Equal[error](nil)),
			Expect(string(raw), Equal("config")),
		)

		manifestDesc, err := courier.DoWith(ctx, storeClient, &endpointstorev1.PutManifest{
			Namespace: ns,
			Digest:    "sha256:manifest",
			Body: &storev1.Manifest{
				Config: *config,
				Assets: []storev1.Descriptor{*asset},
			},
		})
		Then(t, "写入 manifest 接口可返回描述对象",
			Expect(err, Equal[error](nil)),
			Expect(manifestDesc.Digest, Equal(storev1.Digest("sha256:manifest"))),
		)

		manifest, err := courier.DoWith(ctx, storeClient, &endpointstorev1.GetManifest{
			Namespace: ns,
			Digest:    "sha256:manifest",
		})
		Then(t, "读取 manifest 接口可返回完整清单",
			Expect(err, Equal[error](nil)),
			Expect(manifest.Config.Digest, Equal(config.Digest)),
			Expect(len(manifest.Assets), Equal(1)),
			Expect(manifest.Assets[0].Digest, Equal(asset.Digest)),
		)

		_, err = courier.DoWith(ctx, storeClient, &endpointstorev1.DeleteManifest{
			Namespace: ns,
			Digest:    "sha256:manifest",
		})
		Then(t, "删除 manifest 接口可成功执行",
			Expect(err, Equal[error](nil)),
		)

		_, err = courier.DoWith(ctx, storeClient, &endpointstorev1.GetManifest{
			Namespace: ns,
			Digest:    "sha256:manifest",
		})
		Then(t, "删除后再次读取 manifest 会返回不存在错误",
			ExpectDo(func() error { return err }, ErrorAsType[*statuserror.Descriptor]()),
			ExpectMust(func() error { return expectStatusCode(err, http.StatusNotFound) }),
		)

		deleted, err := courier.DoWith(ctx, storeClient, &endpointstorev1.DeleteBlob{
			Namespace: ns,
			Digest:    asset.Digest,
		})
		Then(t, "删除 blob 接口可返回被删对象描述",
			Expect(err, Equal[error](nil)),
			Expect(deleted.Digest, Equal(asset.Digest)),
		)
		_, err = courier.DoWith(ctx, storeClient, &endpointstorev1.GetBlob{
			Namespace: ns,
			Digest:    asset.Digest,
		})

		Then(t, "删除后再次读取 blob 会返回不存在错误",
			ExpectDo(func() error { return err }, ErrorAsType[*statuserror.Descriptor]()),
			ExpectMust(func() error { return expectStatusCode(err, http.StatusNotFound) }),
		)
	})
}

func newTestServer(t *testing.T) *httptest.Server {
	t.Helper()

	h, err := httprouter.New(R, "example")
	if err != nil {
		t.Fatalf("failed to build router: %v", err)
	}

	orgSvc := orgmem.New()
	storeSvc := storemem.New()

	h = handler.ApplyMiddlewares(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			ctx := orgservice.ServiceInjectContext(req.Context(), orgSvc)
			ctx = storeservice.ServiceInjectContext(ctx, storeSvc)
			next.ServeHTTP(rw, req.WithContext(ctx))
		})
	})(h)

	return httptest.NewServer(h)
}

func ioReadCloser(s string) io.ReadCloser {
	return io.NopCloser(bytes.NewBufferString(s))
}

func newOrgClient(srv *httptest.Server) *client.Client {
	return &client.Client{
		Endpoint: srv.URL + "/api/example",
		NewError: func() error { return &statuserror.Descriptor{} },
	}
}

func newStoreClient(srv *httptest.Server) *client.Client {
	return &client.Client{
		Endpoint: srv.URL + "/api/store",
		NewError: func() error { return &statuserror.Descriptor{} },
	}
}

func expectStatusCode(err error, code int) error {
	var desc *statuserror.Descriptor
	if !errors.As(err, &desc) {
		return fmt.Errorf("unexpected error type: %T", err)
	}
	if desc.StatusCode() != code {
		return fmt.Errorf("unexpected status code: %d", desc.StatusCode())
	}
	return nil
}

func ptr[T any](v T) *T {
	return &v
}
