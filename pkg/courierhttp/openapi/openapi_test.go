package openapi_test

import (
	"testing"

	"github.com/octohelm/courier/example/apis"
	"github.com/octohelm/courier/internal/testingutil"
	"github.com/octohelm/courier/pkg/courierhttp/openapi"
)

func TestOpenapi(t *testing.T) {
	o := openapi.FromRouter(apis.R)

	testingutil.PrintJSON(o)
}
