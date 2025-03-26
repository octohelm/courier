// +gengo:operator:tag=blob
// +gengo:operator:register=R
//
//go:generate go tool gen .
package blob

import "github.com/octohelm/courier/pkg/courier"

var R = courier.NewRouter()
