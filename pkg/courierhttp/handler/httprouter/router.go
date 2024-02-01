package httprouter

import (
	"fmt"
	"net/http"
	"os"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/octohelm/courier/internal/request"
	"github.com/octohelm/courier/pkg/courier"
	"github.com/octohelm/courier/pkg/courierhttp/handler"
	"github.com/octohelm/courier/pkg/courierhttp/openapi"
)

type RouteHandler = request.RouteHandler

func NewRouteHandler(route courier.Route, service string) (RouteHandler, error) {
	return request.NewRouteHandler(route, service)
}

func New(cr courier.Router, service string, middlewares ...handler.HandlerMiddleware) (http.Handler, error) {
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
	}

	routes := cr.Routes()

	oas := openapi.DefaultOpenAPIBuildFunc(cr)

	handlers := make([]request.RouteHandler, 0, len(routes))

	for i := range routes {
		rh, err := request.NewRouteHandler(routes[i], service)
		if err != nil {
			return nil, err
		}
		handlers = append(handlers, rh)
	}

	sort.Slice(handlers, func(i, j int) bool {
		return handlers[i].Path() < handlers[j].Path()
	})

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', tabwriter.TabIndent)
	defer func() {
		_ = w.Flush()
	}()

	t := &mux{
		oas:              oas,
		globalMiddleware: handler.ApplyHandlerMiddlewares(append(middlewares, methodOverride)...),
	}

	for i := range handlers {
		h := handlers[i]

		_, _ = fmt.Fprintf(w, "%s\t%s", h.Method()[0:3], h.PathSegments())
		_, _ = fmt.Fprintf(w, "\t%s", h.Summary())
		for _, o := range h.Operators() {
			_, _ = fmt.Fprintf(w, "\t%s", o.String())
		}
		_, _ = fmt.Fprintf(w, "\n")

		t.register(h)
	}

	return t.Handler()
}

var methodOverride = func(n http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if methodOverwrite := req.Header.Get("X-HTTP-Method-Override"); methodOverwrite != "" {
			req.Method = strings.ToUpper(methodOverwrite)
		}
		n.ServeHTTP(rw, req)
	})
}
