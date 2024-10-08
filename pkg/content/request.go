package content

import (
	"context"
	"net/http"

	"github.com/octohelm/courier/internal/httprequest"
	"github.com/octohelm/courier/pkg/content/internal"
)

type RequestInfo = httprequest.Request

func UnmarshalRequestInfo(ireq RequestInfo, out any) error {
	return internal.UnmarshalRequestInfo(ireq, out)
}

func UnmarshalRequest(req *http.Request, out any) error {
	return internal.UnmarshalRequest(req, out)
}

func NewRequest(ctx context.Context, method string, path string, v any) (*http.Request, error) {
	return internal.NewRequest(ctx, method, path, v)
}
