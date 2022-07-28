package httprouter_test

import (
	"net/http"
	"testing"
	"time"

	"github.com/octohelm/courier/example/apis"
	"github.com/octohelm/courier/example/apis/org"
	"github.com/octohelm/courier/internal/testingutil"
	"github.com/octohelm/courier/pkg/courierhttp"
	"github.com/octohelm/courier/pkg/courierhttp/handler/httprouter"
)

func TestNew(t *testing.T) {
	h, err := httprouter.New(apis.R, "test")
	testingutil.Expect(t, err, testingutil.Be[error](nil))

	t.Run("Redirect", func(t *testing.T) {
		type ListOrgOld struct {
			courierhttp.MethodGet `path:"/api/example/v0/org"`
		}

		testingutil.ShouldReturnWhenRequest(t, h, &ListOrgOld{}, `
HTTP/0.0 302 Found
Content-Type: text/html; charset=utf-8
Location: /orgs

<a href="/orgs">Found</a>.
`)
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

		testingutil.ShouldReturnWhenRequest(t, h, &Cookie{
			Token: cookie.Value,
		}, `
HTTP/0.0 204 No Content
Set-Cookie: `+cookie.String()+`

`)
	})

	t.Run("return ok", func(t *testing.T) {
		type GetOrg struct {
			courierhttp.MethodGet `path:"/api/example/v0/orgs/:orgName"`
			Name                  string `name:"orgName" in:"path"`
		}

		testingutil.ShouldReturnWhenRequest(t, h, &GetOrg{
			Name: "hello",
		}, `HTTP/0.0 200 OK
Content-Type: application/json; charset=utf-8

{"name":"hello","type":"GOV"}
`)
	})

	t.Run("POST", func(t *testing.T) {
		type AddOrg struct {
			courierhttp.MethodPost `path:"/api/example/v0/orgs"`
			org.Info               `in:"body"`
		}

		t.Run("return ok", func(t *testing.T) {
			testingutil.ShouldReturnWhenRequest(t, h, &AddOrg{
				Info: org.Info{
					Name: "x",
					Type: org.TYPE__GOV,
				},
			}, `
HTTP/0.0 204 No Content
`)
		})

		t.Run("POST return failed", func(t *testing.T) {
			testingutil.ShouldReturnWhenRequest(t, h, &AddOrg{
				Info: org.Info{
					Name: "xxxxxxx",
				},
			}, `
HTTP/0.0 400 Bad Request
Content-Type: application/json; charset=utf-8

{"code":400,"key":"badRequest","msg":"invalid parameters","desc":"","canBeTalkError":false,"sources":["test"],"errorFields":[{"field":"name","msg":"string length should be less than 5, but got invalid value 7","in":"body"}]}
`)
		})
	})

}
