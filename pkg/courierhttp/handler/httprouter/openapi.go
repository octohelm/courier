package httprouter

import (
	"context"
	"github.com/octohelm/courier/pkg/openapi"

	"github.com/octohelm/courier/pkg/courierhttp"
)

type OpenAPI struct {
	courierhttp.MethodGet
}

func (o *OpenAPI) Output(ctx context.Context) (any, error) {
	return &openapi.Payload{
		OpenAPI: *openapi.FromContext(ctx),
	}, nil
}

func (o *OpenAPI) ResponseContentType() string {
	return "application/json"
}
