// +gengo:operator:register=R
// +gengo:runtimedoc=false
//
// Package v1 组装示例制品仓库域的 v1 路由操作符，并注册到本包导出的 Router。
//
//go:generate go tool gen .
package v1

import (
	"github.com/octohelm/courier/pkg/courier"
)

var R = courier.NewRouter()
