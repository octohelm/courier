package handler_test

import (
	"context"
	"testing"

	"github.com/octohelm/courier/pkg/courier"
	"github.com/octohelm/courier/pkg/courierhttp"
	"github.com/octohelm/courier/pkg/courierhttp/handler/httprouter"
)

type HandlerTestPing struct {
	courierhttp.MethodGet `path:"/ping"`
}

func (*HandlerTestPing) Output(context.Context) (any, error) {
	return "pong", nil
}

func Test(t *testing.T) {
	r := courierhttp.GroupRouter("/").With(
		courier.NewRouter(&HandlerTestPing{}),
	)

	_, _ = httprouter.New(r, "demo")
}
