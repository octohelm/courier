package openapi_test

import (
	"context"
	"testing"

	"github.com/octohelm/courier/internal/testingutil"
	"github.com/octohelm/courier/pkg/courier"
	"github.com/octohelm/courier/pkg/courierhttp"
	"github.com/octohelm/courier/pkg/courierhttp/openapi"
)

type OpenAPITestGetOrg struct {
	courierhttp.MethodGet `path:"/api/example/v1/orgs/{orgID}"`
	OrgID                 uint64 `name:"orgID" in:"path"`
}

func (*OpenAPITestGetOrg) Output(context.Context) (any, error) {
	return struct {
		ID uint64 `json:"id"`
	}{ID: 1}, nil
}

func TestOpenapi(t *testing.T) {
	r := courierhttp.GroupRouter("/").With(
		courier.NewRouter(&OpenAPITestGetOrg{}),
	)

	o := openapi.FromRouter(r)

	testingutil.PrintJSON(o)
}
