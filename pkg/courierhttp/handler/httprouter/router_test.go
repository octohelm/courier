package httprouter_test

import (
	"net/http"
	"testing"
	"time"

	testingx "github.com/octohelm/x/testing"

	"github.com/octohelm/courier/example/apis"
	"github.com/octohelm/courier/example/apis/org"
	"github.com/octohelm/courier/internal/testingutil"
	"github.com/octohelm/courier/pkg/courierhttp"
	"github.com/octohelm/courier/pkg/courierhttp/handler/httprouter"
)

func TestNew(t *testing.T) {
	h, err := httprouter.New(apis.R, "test")
	testingx.Expect(t, err, testingx.BeNil[error]())

	t.Run("Redirect", func(t *testing.T) {
		type ListOrgOld struct {
			courierhttp.MethodGet `path:"/api/example/v0/org"`
		}

		testingx.Expect(t, h, testingutil.ShouldReturnWhenRequest(&ListOrgOld{}, `
HTTP/0.0 302 Found
Content-Type: text/html; charset=utf-8
Location: /orgs
Server: test (ListOrgOld)

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
Server: test (Cookie)
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
Server: test (GetOrg)

{"name":"hello","type":"GOV"}
`))
	})

	t.Run("POST", func(t *testing.T) {
		type AddOrg struct {
			courierhttp.MethodPost `path:"/api/example/v0/orgs"`
			org.Info               `in:"body"`
		}

		t.Run("return 204", func(t *testing.T) {
			testingx.Expect(t,
				h,
				testingutil.ShouldReturnWhenRequest(&AddOrg{
					Info: org.Info{
						Name: "x",
						Type: org.TYPE__GOV,
					},
				}, `
HTTP/0.0 204 No Content
Server: test (CreateOrg)
`))
		})

		t.Run("POST return failed", func(t *testing.T) {
			testingx.Expect(t, h, testingutil.ShouldReturnWhenRequest(&AddOrg{
				Info: org.Info{
					Name: "xxxxxxx",
				},
			}, `
HTTP/0.0 400 Bad Request
Content-Type: application/json; charset=utf-8
Server: test (CreateOrg)

{"code":400,"key":"InvalidParameter","msg":"Bad Request","source":"test","errors":[{"key":"InvalidParameter","msg":"string value length should be less or equal than 5, but got 7","location":"body","pointer":"/name","source":"test"}]}
`))
		})
	})

}
