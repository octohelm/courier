package httprouter_test

import (
	"context"
	"net/http"
	"net/url"
	"regexp"
	"testing"
	"time"

	"github.com/octohelm/courier/internal/testingutil"
	"github.com/octohelm/courier/pkg/courier"
	"github.com/octohelm/courier/pkg/courierhttp"
	"github.com/octohelm/courier/pkg/courierhttp/handler/httprouter"
	"github.com/octohelm/courier/pkg/validator"
	"github.com/octohelm/courier/pkg/validator/validators"
	testingx "github.com/octohelm/x/testing"
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
	testingx.Expect(t, err, testingx.BeNil[error]())

	t.Run("Redirect", func(t *testing.T) {
		type ListOrgOld struct {
			courierhttp.MethodGet `path:"/api/example/v0/org"`
		}

		testingx.Expect(t, h, testingutil.ShouldReturnWhenRequest(&ListOrgOld{}, `
HTTP/0.0 302 Found
Content-Type: text/html; charset=utf-8
Location: /orgs
Server: test (testRouterListOrgOld)

<a href="/orgs">Found</a>.
`))
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

		testingx.Expect(t, h, testingutil.ShouldReturnWhenRequest(&Cookie{
			Token: cookie.Value,
		}, `
HTTP/0.0 204 No Content
Server: test (testRouterCookie)
Set-Cookie: `+cookie.String()+`

`))
	})

	t.Run("return ok", func(t *testing.T) {
		type GetOrg struct {
			courierhttp.MethodGet `path:"/api/example/v0/orgs/{orgName}"`
			Name                  string `name:"orgName" in:"path"`
		}

		testingx.Expect(t, h, testingutil.ShouldReturnWhenRequest(&GetOrg{
			Name: "hello",
		}, `HTTP/0.0 200 OK
Content-Type: application/json; charset=utf-8
Server: test (testRouterGetOrg)

{"name":"hello","type":"GOV"}
`))
	})

	t.Run("POST", func(t *testing.T) {
		type AddOrg struct {
			courierhttp.MethodPost `path:"/api/example/v0/orgs"`
			TestOrgInfo            `in:"body"`
		}

		t.Run("return 204", func(t *testing.T) {
			testingx.Expect(t,
				h,
				testingutil.ShouldReturnWhenRequest(&AddOrg{
					TestOrgInfo: TestOrgInfo{
						Name: "x",
						Type: testOrgTypeGov,
					},
				}, `
HTTP/0.0 204 No Content
Server: test (testRouterCreateOrg)
`))
		})

		t.Run("POST return failed", func(t *testing.T) {
			testingx.Expect(t, h, testingutil.ShouldReturnWhenRequest(&AddOrg{
				TestOrgInfo: TestOrgInfo{
					Name: "xxxxxxx",
				},
			}, `
HTTP/0.0 400 Bad Request
Content-Type: application/json; charset=utf-8
Server: test (testRouterCreateOrg)

{"code":400,"msg":"Bad Request","errors":[{"code":"INVALID_PARAMETER","message":"string value length should be less or equal than 5, but got 7","location":"body","pointer":"/name","source":"test"}]}
`))
		})
	})
}
