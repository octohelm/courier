package httprouter

import (
	"context"
	"fmt"

	"github.com/octohelm/courier/pkg/courierhttp"
	"github.com/octohelm/courier/pkg/openapi"
	"github.com/octohelm/courier/pkg/statuserror"
)

var forbiddenOpenAPI = false

func ForbidOpenAPI(forbidden bool) {
	forbiddenOpenAPI = forbidden
}

type ErrOpenAPIForbidden struct {
	statuserror.Forbidden
}

func (e *ErrOpenAPIForbidden) Error() string {
	return fmt.Sprintf("openapi is forbidden")
}

type OpenAPI struct {
	courierhttp.MethodGet
}

func (o *OpenAPI) Output(ctx context.Context) (any, error) {
	if forbiddenOpenAPI {
		return nil, &ErrOpenAPIForbidden{}
	}

	if x, ok := courierhttp.OperationInfoProviderFromContext(ctx); ok {
		if o, ok := x.(interface{ OpenAPI() *openapi.OpenAPI }); ok {
			return &openapi.Payload{
				OpenAPI: *o.OpenAPI(),
			}, nil

		}
	}

	return nil, &ErrOpenAPIForbidden{}
}

func (o *OpenAPI) ResponseContentType() string {
	return "application/json"
}
