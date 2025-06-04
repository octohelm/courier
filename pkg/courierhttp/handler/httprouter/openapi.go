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

	return &openapi.Payload{
		OpenAPI: *openapi.FromContext(ctx),
	}, nil
}

func (o *OpenAPI) ResponseContentType() string {
	return "application/json"
}
