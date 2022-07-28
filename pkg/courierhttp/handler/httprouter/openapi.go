package httprouter

import (
	"context"

	"github.com/octohelm/courier/pkg/courierhttp"
	"github.com/octohelm/courier/pkg/openapi"
)

var oas = openapi.NewOpenAPI()

type OpenAPI struct {
	courierhttp.MethodGet
}

func (o *OpenAPI) Output(ctx context.Context) (any, error) {
	return oas, nil
}

func (o *OpenAPI) ResponseContentType() string {
	return "application/json"
}
