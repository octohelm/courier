package transport

import (
	"context"
	"net/http"

	"github.com/go-courier/logr"
	"github.com/octohelm/courier/pkg/content"

	"github.com/octohelm/courier/pkg/courierhttp"
)

type IncomingTransport interface {
	UnmarshalOperator(ctx context.Context, info courierhttp.Request, op any) error
	WriteResponse(ctx context.Context, rw http.ResponseWriter, result any, info courierhttp.Request)
}

func NewIncomingTransport(ctx context.Context, v any) (IncomingTransport, error) {

	return &incomingTransport{}, nil
}

type incomingTransport struct {
}

func (t *incomingTransport) UnmarshalOperator(ctx context.Context, ireq courierhttp.Request, op any) error {
	return content.UnmarshalRequestInfo(ireq, op)
}

func (i *incomingTransport) WriteResponse(ctx context.Context, rw http.ResponseWriter, ret any, req courierhttp.Request) {
	if upgrader, ok := ret.(Upgrader); ok {
		if err := upgrader.Upgrade(rw, req.Underlying()); err != nil {
			i.writeErrResp(ctx, rw, err, req)
		}
		return
	}

	if err, ok := ret.(error); ok {
		i.writeErrResp(ctx, rw, err, req)
	} else {
		i.writeResp(ctx, rw, ret, req)
	}
}

func (i *incomingTransport) writeResp(ctx context.Context, rw http.ResponseWriter, ret any, req courierhttp.Request) {
	if err := courierhttp.Wrap(ret).(courierhttp.ResponseWriter).WriteResponse(ctx, rw, req); err != nil {
		logr.FromContext(ctx).Error(err)
	}
}

func (i *incomingTransport) writeErrResp(ctx context.Context, rw http.ResponseWriter, err error, req courierhttp.Request) {
	if err := courierhttp.WrapError(err).(courierhttp.ResponseWriter).WriteResponse(ctx, rw, req); err != nil {
		logr.FromContext(ctx).Error(err)
	}
}
