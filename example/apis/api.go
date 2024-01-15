package apis

import (
	"github.com/octohelm/courier/example/apis/blob"
	"github.com/octohelm/courier/example/apis/org"
	"github.com/octohelm/courier/pkg/courier"
	"github.com/octohelm/courier/pkg/courierhttp"

	"github.com/octohelm/courier/pkg/courierhttp/handler/httprouter"
)

var R = courierhttp.GroupRouter("/api/example").With(
	courier.NewRouter(&httprouter.OpenAPI{}),
	courierhttp.GroupRouter("/v0").With(
		org.R,
		blob.R,
	),
)
