package httprouter

import (
	"net/http"
	"sort"
	"strings"

	"github.com/octohelm/courier/internal/request"
	"github.com/octohelm/courier/pkg/courier"
	"github.com/octohelm/courier/pkg/courierhttp/handler"
	"github.com/octohelm/courier/pkg/courierhttp/openapi"
)

type RouteHandler = request.RouteHandler

func NewRouteHandlers(route courier.Route, service string, routeMiddlewares ...handler.Middleware) ([]RouteHandler, error) {
	return request.NewRouteHandlers(route, service, routeMiddlewares...)
}

func New(cr courier.Router, service string, routeMiddlewares ...handler.Middleware) (http.Handler, error) {
	customOpenApiRouter := false

	for _, r := range cr.Routes() {
		if customOpenApiRouter {
			break
		}

		_ = r.RangeOperator(func(f *courier.OperatorFactory, i int) error {
			if f.IsLast {
				if _, ok := f.Operator.(*OpenAPI); ok {
					customOpenApiRouter = true
				}
			}
			return nil
		})
	}

	if !customOpenApiRouter {
		cr.Register(courier.NewRouter(&OpenAPI{}))
		cr.Register(courier.NewRouter(&OpenAPIView{}))
	}

	routes := cr.Routes()

	oas := openapi.DefaultBuildFunc(cr)
	oas.Title = service

	handlers := make([]request.RouteHandler, 0, len(routes))

	for i := range routes {
		// middleware for each route
		rh, err := NewRouteHandlers(routes[i], service, routeMiddlewares...)
		if err != nil {
			return nil, err
		}
		handlers = append(handlers, rh...)
	}

	sort.Slice(handlers, func(i, j int) bool {
		return handlers[i].Path() < handlers[j].Path()
	})

	m := &mux{
		oas: oas,
	}

	nameVersion := strings.Split(service, "@")

	m.server.Name = nameVersion[0]
	if len(nameVersion) >= 2 {
		m.server.Version = nameVersion[1]
	}

	for i := range handlers {
		m.register(handlers[i])
	}

	h, err := m.Handler()
	if err != nil {
		return nil, err
	}

	return methodOverride(h), nil
}

var methodOverride = func(n http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if methodOverwrite := req.Header.Get("X-HTTP-Method-Override"); methodOverwrite != "" {
			req.Method = strings.ToUpper(methodOverwrite)
		}
		n.ServeHTTP(rw, req)
	})
}
