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

type MethodDescriber interface {
	Method() string
}

type PathDescriber interface {
	Path() string
}

type BasePathDescriber interface {
	BasePath() string
}

func GroupRouter(p string) courier.Router {
	return courier.NewRouter(Group(p))
}

func BasePathRouter(basePath string) courier.Router {
	return courier.NewRouter(BasePath(basePath))
}

func BasePath(basePath string) courier.Operator {
	return &metaOperator{basePath: path.Clean(basePath)}
}

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
