package handler_test

import (
	"testing"

	"github.com/octohelm/courier/example/apis"
	"github.com/octohelm/courier/pkg/courierhttp/handler/httprouter"
)

func Test(t *testing.T) {
	_, _ = httprouter.New(apis.R, "demo")
}
