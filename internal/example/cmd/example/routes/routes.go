//go:generate go tool gen .
package routes

import (
	orgv1 "github.com/octohelm/courier/internal/example/cmd/example/routes/org/v1"
	storev1 "github.com/octohelm/courier/internal/example/cmd/example/routes/store/v1"
	"github.com/octohelm/courier/pkg/courier"
	"github.com/octohelm/courier/pkg/courierhttp"
	"github.com/octohelm/courier/pkg/courierhttp/handler/httprouter"
)

var R = courierhttp.GroupRouter("/api").With(
	courier.NewRouter(&httprouter.OpenAPI{}),
	courierhttp.GroupRouter("/example").With(
		orgv1.R,
	),
	courierhttp.GroupRouter("/store").With(
		storev1.R,
	),
)
