// +gengo:operator:tag=store
// +gengo:operator:register=R
//
//go:generate go tool gen .
package store

import (
	"github.com/octohelm/courier/pkg/courierhttp"
)

var R = courierhttp.GroupRouter("/store")
