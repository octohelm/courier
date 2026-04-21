package mem

import (
	"context"
	"testing"

	metav1 "github.com/octohelm/courier/internal/example/pkg/apis/meta/v1"
	orgv1 "github.com/octohelm/courier/internal/example/pkg/apis/org/v1"
	. "github.com/octohelm/x/testing/v2"
)

func TestService(t *testing.T) {
	svc := New()
	ctx := context.Background()

	t.Run("创建组织并查询列表", func(t *testing.T) {
		org, err := svc.Create(ctx, &orgv1.OrgForCreateRequest{
			Spec: orgv1.OrgSpec{
				Name: "alpha",
				Type: orgv1.ORG_TYPE__GOV,
			},
		})
		Then(t, "创建组织后可命中名称过滤",
			Expect(err, Equal[error](nil)),
			Expect(org.ID > 0, Equal(true)),
		)

		list, err := svc.List(ctx, &orgv1.OrgForListRequest{
			OrgName: ptr(orgv1.OrgName("alpha")),
		}, &metav1.Pager{})
		Then(t, "列表查询结果符合预期",
			Expect(err, Equal[error](nil)),
			Expect(list.Total, Equal(int64(1))),
			Expect(len(list.Items), Equal(1)),
		)
	})

	t.Run("组织名冲突时报错", func(t *testing.T) {
		_, err := svc.Create(ctx, &orgv1.OrgForCreateRequest{
			Spec: orgv1.OrgSpec{
				Name: "alpha",
				Type: orgv1.ORG_TYPE__COMPANY,
			},
		})
		Then(t, "重复创建同名组织会返回冲突错误",
			ExpectDo(func() error { return err }, ErrorAsType[*orgv1.ErrOrgNameConflict]()),
		)
	})

	t.Run("更新和删除组织", func(t *testing.T) {
		org, err := svc.Create(ctx, &orgv1.OrgForCreateRequest{
			Spec: orgv1.OrgSpec{
				Name: "beta",
				Type: orgv1.ORG_TYPE__COMPANY,
			},
		})
		Then(t, "准备更新数据成功",
			Expect(err, Equal[error](nil)),
		)

		updated, err := svc.Update(ctx, org.ID, &orgv1.OrgForUpdateRequest{
			Spec: orgv1.OrgSpec{
				Name: "beta2",
				Type: orgv1.ORG_TYPE__GOV,
			},
		})
		Then(t, "更新组织后名称发生变化",
			Expect(err, Equal[error](nil)),
			Expect(updated.Spec.Name, Equal(orgv1.OrgName("beta2"))),
		)

		err = svc.Delete(ctx, org.ID)
		Then(t, "删除组织成功",
			Expect(err, Equal[error](nil)),
		)

		_, err = svc.Get(ctx, org.ID)
		Then(t, "删除后再次查询会返回不存在错误",
			ExpectDo(func() error { return err }, ErrorAsType[*orgv1.ErrOrgNotFound]()),
		)
	})
}

func ptr[T any](v T) *T {
	return &v
}
