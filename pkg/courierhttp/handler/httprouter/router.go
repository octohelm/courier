package httprouter

import (
	"fmt"
	"net/http"
	"os"
	"regexp"
	"sort"
	"text/tabwriter"

	"github.com/julienschmidt/httprouter"
	"github.com/octohelm/courier/internal/request"
	"github.com/octohelm/courier/pkg/courier"
	"github.com/octohelm/courier/pkg/courierhttp"
	"github.com/octohelm/courier/pkg/courierhttp/handler"
	"github.com/octohelm/courier/pkg/courierhttp/openapi"
)

func New(cr courier.Router, service string, middlewares ...handler.HandlerMiddleware) (http.Handler, error) {
	r := httprouter.New()

	cr.Register(courier.NewRouter(&OpenAPI{}))

	oas = openapi.DefaultOpenAPIBuildFunc(cr)

	routes := cr.Routes()

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

		hh := handler.ApplyHandlerMiddlewares(middlewares...)(h)

		_, _ = fmt.Fprintf(w, "%s\t%s", h.Method()[0:3], reHttpRouterPath.ReplaceAllString(h.Path(), "/{$1}"))
		_, _ = fmt.Fprintf(w, "\t%s", h.Summary())
		for _, o := range h.Operators() {
			_, _ = fmt.Fprintf(w, "\t%s", o.String())
		}
		_, _ = fmt.Fprintf(w, "\n")

		r.Handle(h.Method(), h.Path(), func(rw http.ResponseWriter, req *http.Request, params httprouter.Params) {
			ctx := req.Context()
			ctx = courierhttp.ContextWithOperationID(ctx, h.OperationID())
			ctx = handler.ContextWithParamGetter(ctx, params)
			hh.ServeHTTP(rw, req.WithContext(ctx))
		})
	}

	return r, nil
}

var reHttpRouterPath = regexp.MustCompile("/:([^/]+)")
