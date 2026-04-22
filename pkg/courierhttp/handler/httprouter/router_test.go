package httprouter_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"testing"
	"time"

	. "github.com/octohelm/x/testing/v2"

	"github.com/octohelm/courier/internal/testingutil"
	"github.com/octohelm/courier/pkg/courier"
	"github.com/octohelm/courier/pkg/courierhttp"
	"github.com/octohelm/courier/pkg/courierhttp/handler/httprouter"
	"github.com/octohelm/courier/pkg/validator"
	"github.com/octohelm/courier/pkg/validator/validators"
)

type testOrgType string

const testOrgTypeGov testOrgType = "GOV"

type testOrgName string

func (testOrgName) StructTagValidate() string {
	return "@test-org-name"
}

type TestOrgInfo struct {
	Name testOrgName `json:"name"`
	Type testOrgType `json:"type,omitzero"`
}

type testRouterCookie struct {
	courierhttp.MethodPost `path:"/api/example/v0/cookie-ping-pong"`
	Token                  string `name:"token,omitempty" in:"cookie"`
}

func (req *testRouterCookie) Output(ctx context.Context) (any, error) {
	return courierhttp.Wrap[any](
		nil,
		courierhttp.WithCookies(&http.Cookie{
			Name:    "token",
			Value:   req.Token,
			Expires: time.Now().Add(24 * time.Hour),
		}),
	), nil
}

type testRouterCreateOrg struct {
	courierhttp.MethodPost `path:"/api/example/v0/orgs"`

	TestOrgInfo `in:"body"`
}

func (*testRouterCreateOrg) Output(context.Context) (any, error) {
	return nil, nil
}

type testRouterGetOrg struct {
	courierhttp.MethodGet `path:"/api/example/v0/orgs/{orgName}"`
	Name                  string `name:"orgName" in:"path"`
}

func (req *testRouterGetOrg) Output(context.Context) (any, error) {
	return &TestOrgInfo{
		Name: testOrgName(req.Name),
		Type: testOrgTypeGov,
	}, nil
}

type testRouterListOrgOld struct {
	courierhttp.MethodGet `path:"/api/example/v0/org"`
}

func (*testRouterListOrgOld) Output(context.Context) (any, error) {
	return courierhttp.Redirect(http.StatusFound, &url.URL{Path: "/orgs"}), nil
}

func init() {
	validator.Register(validator.NewFormatValidatorProvider("test-org-name", func(format string) validator.Validator {
		return &validators.StringValidator{
			Format:        format,
			MaxLength:     new(uint64(5)),
			Pattern:       regexp.MustCompile(`^[a-z0-9]([-a-z0-9]*[a-z0-9])?$`),
			PatternErrMsg: "只能包含小写字母，数字和 -，且必须以小写字母或数字开头",
		}
	}))
}

func TestNew(t *testing.T) {
	r := courierhttp.GroupRouter("/").With(
		courier.NewRouter(&testRouterCookie{}),
		courier.NewRouter(&testRouterCreateOrg{}),
		courier.NewRouter(&testRouterGetOrg{}),
		courier.NewRouter(&testRouterListOrgOld{}),
	)

	h, err := httprouter.New(r, "test")
	Then(t, "构建 httprouter handler 成功", Expect(err, Equal[error](nil)))

	t.Run("Redirect", func(t *testing.T) {
		type ListOrgOld struct {
			courierhttp.MethodGet `path:"/api/example/v0/org"`
		}

		Then(t, "旧路由会返回重定向响应",
			Expect(h, Be(testingutil.ShouldReturnWhenRequest(&ListOrgOld{}, `
HTTP/0.0 302 Found
Content-Type: text/html; charset=utf-8
Location: /orgs
Server: test (testRouterListOrgOld)

<a href="/orgs">Found</a>.
`))),
		)
	})

	t.Run("Set-Cookie", func(t *testing.T) {
		type Cookie struct {
			courierhttp.MethodPost `path:"/api/example/v0/cookie-ping-pong"`
			Token                  string `name:"token" in:"cookie"`
		}

		cookie := &http.Cookie{
			Name:    "token",
			Value:   "test",
			Expires: time.Now().Add(24 * time.Hour),
		}

		Then(t, "cookie 会被正确写回响应头",
			Expect(h, Be(testingutil.ShouldReturnWhenRequest(&Cookie{
				Token: cookie.Value,
			}, `
HTTP/0.0 204 No Content
Server: test (testRouterCookie)
Set-Cookie: `+cookie.String()+`

`))),
		)
	})

	t.Run("return ok", func(t *testing.T) {
		type GetOrg struct {
			courierhttp.MethodGet `path:"/api/example/v0/orgs/{orgName}"`
			Name                  string `name:"orgName" in:"path"`
		}

		Then(t, "GET 请求会返回 JSON 响应",
			Expect(h, Be(testingutil.ShouldReturnWhenRequest(&GetOrg{
				Name: "hello",
			}, `HTTP/0.0 200 OK
Content-Type: application/json; charset=utf-8
Server: test (testRouterGetOrg)

{"name":"hello","type":"GOV"}
`))),
		)
	})

	t.Run("POST routes", func(t *testing.T) {
		type AddOrg struct {
			courierhttp.MethodPost `path:"/api/example/v0/orgs"`
			TestOrgInfo            `in:"body"`
		}

		t.Run("returns 204 on valid request", func(t *testing.T) {
			Then(t, "合法请求会返回 204", Expect(h, Be(testingutil.ShouldReturnWhenRequest(&AddOrg{
				TestOrgInfo: TestOrgInfo{
					Name: "x",
					Type: testOrgTypeGov,
				},
			}, `
HTTP/0.0 204 No Content
Server: test (testRouterCreateOrg)
`))))
		})

		t.Run("returns validation error on invalid request", func(t *testing.T) {
			Then(t, "非法请求会返回带 pointer 的校验错误", Expect(h, Be(testingutil.ShouldReturnWhenRequest(&AddOrg{
				TestOrgInfo: TestOrgInfo{
					Name: "xxxxxxx",
				},
			}, `
HTTP/0.0 400 Bad Request
Content-Type: application/json; charset=utf-8
Server: test (testRouterCreateOrg)

{"code":400,"msg":"Bad Request","errors":[{"code":"INVALID_PARAMETER","message":"string value length should be less or equal than 5, but got 7","location":"body","pointer":"/name","source":"test"}]}
`))))
		})
	})
}

func TestRouteSnapshot(t *testing.T) {
	r := courierhttp.GroupRouter("/").With(
		courier.NewRouter(&testRouterCookie{}),
		courier.NewRouter(&testRouterCreateOrg{}),
		courier.NewRouter(&testRouterGetOrg{}),
		courier.NewRouter(&testRouterListOrgOld{}),
	)

	Then(t, "路由快照会包含关键 method、path 和 operator 链", ExpectMust(func() error {
		snapshot, err := httprouter.RouteSnapshot(r, "test")
		if err != nil {
			return err
		}
		for _, fragment := range []string{
			"GET   /api/example/v0/org",
			"GET   /api/example/v0/orgs/{orgName}",
			"POST  /api/example/v0/orgs",
			"{{ httprouter_test.testRouterGetOrg }}",
			"{{ httprouter.OpenAPI }}",
		} {
			if !strings.Contains(snapshot, fragment) {
				return fmt.Errorf("snapshot missing fragment: %s", fragment)
			}
		}
		return nil
	}))
}

func TestRouteSnapshotMatchesStartupOutput(t *testing.T) {
	r := courierhttp.GroupRouter("/").With(
		courier.NewRouter(&testRouterCookie{}),
		courier.NewRouter(&testRouterCreateOrg{}),
		courier.NewRouter(&testRouterGetOrg{}),
		courier.NewRouter(&testRouterListOrgOld{}),
	)

	snapshot, err := httprouter.RouteSnapshot(r, "test")
	Then(t, "路由快照构建成功", Expect(err, Equal[error](nil)))

	originStdout := os.Stdout
	reader, writer, err := os.Pipe()
	Then(t, "可以捕获启动时标准输出", Expect(err, Equal[error](nil)))
	os.Stdout = writer
	defer func() {
		os.Stdout = originStdout
	}()

	_, err = httprouter.New(r, "test")
	Then(t, "构建 handler 时会打印启动快照", Expect(err, Equal[error](nil)))

	_ = writer.Close()

	data, err := io.ReadAll(reader)
	Then(t, "可以读取捕获到的输出", Expect(err, Equal[error](nil)))

	startupOutput := strings.TrimSpace(bytes.NewBuffer(data).String())
	Then(t, "启动输出与 RouteSnapshot 结果一致", Expect(startupOutput, Equal(snapshot)))
}
