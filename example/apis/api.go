package apis

import (
	"github.com/octohelm/courier/example/apis/blob"
	"github.com/octohelm/courier/example/apis/org"
	"github.com/octohelm/courier/pkg/courierhttp"
)

var R = courierhttp.GroupRouter("/api/example").With(
	courierhttp.GroupRouter("/v0").With(
		org.R,
		blob.R,
	),
)
