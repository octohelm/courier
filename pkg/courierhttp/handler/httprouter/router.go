package httprouter

import (
	"fmt"
	"net/http"
	"os"
	"regexp"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/julienschmidt/httprouter"
	"github.com/octohelm/courier/internal/request"
	"github.com/octohelm/courier/pkg/courier"
	"github.com/octohelm/courier/pkg/courierhttp"
	"github.com/octohelm/courier/pkg/courierhttp/handler"
	"github.com/octohelm/courier/pkg/courierhttp/openapi"
)

type RouteHandler = request.RouteHandler

func NewRouteHandler(route courier.Route, service string) (RouteHandler, error) {
	return request.NewRouteHandler(route, service)
}

func New(cr courier.Router, service string, middlewares ...handler.HandlerMiddleware) (http.Handler, error) {
	r := httprouter.New()

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

	oas = openapi.DefaultOpenAPIBuildFunc(cr)

	handlers := make([]request.RouteHandler, 0, len(routes))

	for i := range routes {
		rh, err := request.NewRouteHandler(routes[i], service)
		if err != nil {
			return nil, err
		}

		method := rh.Method()
		if method == "" {
			continue
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

	for i := range handlers {
		h := handlers[i]

		hh := handler.ApplyHandlerMiddlewares(append(middlewares, methodOverride)...)(h)

		_, _ = fmt.Fprintf(w, "%s\t%s", h.Method()[0:3], reHttpRouterPath.ReplaceAllString(h.Path(), "/{$1}"))
		_, _ = fmt.Fprintf(w, "\t%s", h.Summary())
		for _, o := range h.Operators() {
			_, _ = fmt.Fprintf(w, "\t%s", o.String())
		}
		_, _ = fmt.Fprintf(w, "\n")

		info := courierhttp.OperationInfo{
			ID:     h.OperationID(),
			Method: h.Method(),
			Route:  h.Path(),
		}

		methods := []string{h.Method()}

		if h.Method() == "ALL" {
			methods = []string{
				http.MethodGet,
				http.MethodHead,
				http.MethodPost,
				http.MethodPut,
				http.MethodDelete,
				http.MethodPatch,
				http.MethodConnect,
				http.MethodOptions,
				http.MethodTrace,
			}
		}

		for _, m := range methods {
			r.Handle(m, h.Path(), func(rw http.ResponseWriter, req *http.Request, params httprouter.Params) {
				ctx := req.Context()
				ctx = courierhttp.ContextWithOperationInfo(ctx, info)
				ctx = handler.ContextWithParamGetter(ctx, params)
				hh.ServeHTTP(rw, req.WithContext(ctx))
			})
		}
	}

	return r, nil
}

var reHttpRouterPath = regexp.MustCompile("/:([^/]+)")

var methodOverride = func(n http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if methodOverwrite := req.Header.Get("X-HTTP-Method-Override"); methodOverwrite != "" {
			req.Method = strings.ToUpper(methodOverwrite)
		}
		n.ServeHTTP(rw, req)
	})
}
