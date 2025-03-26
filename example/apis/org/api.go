// +gengo:operator:register=R
// +gengo:operator:tag=org
//
//go:generate go tool gen .
package org

import "github.com/octohelm/courier/pkg/courier"

var R = courier.NewRouter()
