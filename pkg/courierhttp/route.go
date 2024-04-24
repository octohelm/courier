package courierhttp

import (
	"context"
	"fmt"

	contextx "github.com/octohelm/x/context"

	"github.com/octohelm/courier/pkg/courier"
)

type MethodDescriber interface {
	Method() string
}

type PathDescriber interface {
	Path() string
}

type BasePathDescriber interface {
	BasePath() string
}

func GroupRouter(path string) courier.Router {
	return courier.NewRouter(Group(path))
}

func BasePathRouter(basePath string) courier.Router {
	return courier.NewRouter(BasePath(basePath))
}

func BasePath(basePath string) courier.Operator {
	return &metaOperator{basePath: basePath}
}

func Group(path string) courier.Operator {
	return &metaOperator{path: path}
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

type RouteDescriber interface {
	MethodDescriber
	PathDescriber
}

type contextRouteDescriber struct{}

func ContextWithRouteDescriber(ctx context.Context, r RouteDescriber) context.Context {
	return contextx.WithValue(ctx, contextRouteDescriber{}, r)
}

func RouteDescriberFromContext(ctx context.Context) RouteDescriber {
	if v, ok := ctx.Value(contextRouteDescriber{}).(RouteDescriber); ok {
		return v
	}
	return nil
}
