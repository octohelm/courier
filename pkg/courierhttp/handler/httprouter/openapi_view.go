package httprouter

import (
	"context"
	"net/http"

	"github.com/octohelm/courier/pkg/courierhttp"
	"github.com/octohelm/courier/pkg/courierhttp/transport"
	"github.com/octohelm/courier/pkg/statuserror"
	"github.com/pkg/errors"
)

var openapiView transport.Upgrader

func SetOpenAPIViewContents(u transport.Upgrader) {
	openapiView = u
}

type OpenAPIView struct {
	courierhttp.MethodGet `path:"/_view/{href...}"`
}

func (o *OpenAPIView) Output(ctx context.Context) (any, error) {
	if openapiView == nil {
		return nil, statuserror.Wrap(
			errors.New("openapi view not found"),
			http.StatusNotFound,
			"OPENAPI_VIEW_NOT_SUPPORTED",
		)
	}

	return openapiView, nil
}
