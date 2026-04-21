package courierhttp

import (
	"fmt"
	"path"

	"github.com/octohelm/courier/pkg/courier"
)

// +gengo:injectable:provider
type RouteDescriber interface {
	MethodDescriber
	PathDescriber
}

// MethodDescriber 用于描述HTTP方法。
type MethodDescriber interface {
	Method() string
}

// PathDescriber 用于描述路由路径。
type PathDescriber interface {
	Path() string
}

// BasePathDescriber 用于描述基础路径。
type BasePathDescriber interface {
	BasePath() string
}

// GroupRouter 创建带有路径分组的路由器。
// Group 创建路径分组操作符。
func GroupRouter(p string) courier.Router {
	return courier.NewRouter(Group(p))
}

// BasePathRouter 创建带有基础路径的路由器。
// BasePath 创建基础路径操作符。
func BasePathRouter(basePath string) courier.Router {
	return courier.NewRouter(BasePath(basePath))
}

// BasePath 创建基础路径操作符。
func BasePath(basePath string) courier.Operator {
	return &metaOperator{basePath: path.Clean(basePath)}
}

// Group 创建路径分组操作符。
func Group(p string) courier.Operator {
	return &metaOperator{path: path.Clean(p)}
}

type metaOperator struct {
	path     string
	basePath string
	courier.EmptyOperator
}

func (g *metaOperator) Path() string {
	return g.path
}

func (g *metaOperator) BasePath() string {
	return g.basePath
}

func (g *metaOperator) String() string {
	if g.basePath != "" {
		return fmt.Sprintf("basePath(%s)", g.basePath)
	}
	return fmt.Sprintf("group(%s)", g.path)
}
