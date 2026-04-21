// +gengo:operator:register=R
// +gengo:runtimedoc=false
//
//go:generate go tool gen .
package v1

import "github.com/octohelm/courier/pkg/courier"

var R = courier.NewRouter()
