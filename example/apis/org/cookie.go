package org

import (
	"context"
	"net/http"
	"time"

	"github.com/octohelm/courier/pkg/courierhttp"
)

type Cookie struct {
	courierhttp.MethodPost `path:"/cookie-ping-pong"`
	Token                  string `name:"token,omitempty" in:"cookie"`
}

func (req *Cookie) Output(ctx context.Context) (interface{}, error) {
	return courierhttp.Wrap[any](
		nil,
		courierhttp.WithCookies(&http.Cookie{
			Name:    "token",
			Value:   req.Token,
			Expires: time.Now().Add(24 * time.Hour),
		}),
	), nil
}
