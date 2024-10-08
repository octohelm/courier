package request

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/octohelm/courier/internal/httprequest"

	"github.com/octohelm/courier/internal/pathpattern"
	"github.com/octohelm/courier/pkg/courier"
	"github.com/octohelm/courier/pkg/courierhttp"
	"github.com/octohelm/courier/pkg/courierhttp/handler"
	"github.com/octohelm/courier/pkg/courierhttp/transport"
	contextx "github.com/octohelm/x/context"
)

type Segments = pathpattern.Segments

type RouteHandler interface {
	http.Handler

	OperationID() string
	Method() string
	Path() string
	PathSegments() Segments

	Summary() string
	Description() string
	Deprecated() bool

	Operators() []*courier.OperatorFactory
}

func NewRouteHandlers(route courier.Route, service string, routeMiddlewares ...handler.Middleware) ([]RouteHandler, error) {
	h := &routeHandler{
		service:    service,
		middleware: handler.ApplyMiddlewares(routeMiddlewares...),
	}

	basePath := "/"
	path := ""

	err := route.RangeOperator(func(f *courier.OperatorFactory, i int) error {
		m := metaFrom(f)

		if m.BasePath != "" {
			basePath = m.BasePath
		}

		if m.Path != "" {
			path += m.Path
		}

		if f.IsLast {
			h.operationID = f.Type.Name()
			h.deprecated = m.Deprecated
			h.summary = m.Summary
			h.description = m.Description
			if m.Method != "" {
				h.method = m.Method
			}
		}

		if f.NoOutput {
			return nil
		}

		tt, err := transport.NewIncomingTransport(context.Background(), f.New())
		if err != nil {
			return err
		}

		h.operators = append(h.operators, f)
		h.transformers = append(h.transformers, tt)

		return nil
	})
	if err != nil {
		return nil, err
	}

	h.segments = pathpattern.Parse(pathpattern.NormalizePath(basePath + path))

	methods := strings.Split(h.method, ",")

	handlers := make([]RouteHandler, 0, len(methods))

	for _, m := range methods {
		if m == "" {
			continue
		}

		if h.method == m {
			handlers = append(handlers, h)
		} else {
			handlers = append(handlers, h.cloneWithMethod(m))
		}
	}

	return handlers, nil
}

type routeHandler struct {
	service      string
	operationID  string
	method       string
	segments     pathpattern.Segments
	summary      string
	deprecated   bool
	description  string
	operators    []*courier.OperatorFactory
	transformers []transport.IncomingTransport
	middleware   handler.Middleware

	once         sync.Once
	finalHandler http.Handler
}

func (h *routeHandler) OperationID() string {
	return h.operationID
}

func (h *routeHandler) Method() string {
	return h.method
}

func (h *routeHandler) Path() string {
	return h.segments.String()
}

func (h *routeHandler) PathSegments() Segments {
	return h.segments
}

func (h *routeHandler) Summary() string {
	if h.summary == "" {
		return h.OperationID()
	}
	return h.summary
}

func (h *routeHandler) Description() string {
	return h.description
}

func (h *routeHandler) Deprecated() bool {
	return h.deprecated
}

func (h *routeHandler) Operators() []*courier.OperatorFactory {
	return h.operators
}

func (h routeHandler) cloneWithMethod(m string) RouteHandler {
	h.method = m
	h.operationID = fmt.Sprintf("%s_%s", m, h.operationID)
	return &h
}

func (h *routeHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	h.once.Do(func() {
		var hh http.Handler = &routeHttpHandler{
			routeHandler: h,
		}

		if h.middleware != nil {
			hh = h.middleware(hh)
		}

		for _, o := range h.operators {
			if x, ok := o.Operator.(WithPreHandlerMiddleware); ok {
				hh = x.PreHandlerMiddleware(hh)
			}
		}

		h.finalHandler = hh
	})
	h.finalHandler.ServeHTTP(rw, r)
}

type routeHttpHandler struct {
	*routeHandler
}

func (h *routeHttpHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	ctx = courierhttp.HttpRequestInjectContext(ctx, &courierhttp.HttpRequest{Request: r})

	info := httprequest.From(r)

	for i := range h.operators {
		opFactory := h.operators[i]
		t := h.transformers[i]

		op := opFactory.New()

		if err := t.UnmarshalOperator(ctx, info, op); err != nil {
			t.WriteResponse(ctx, rw, err, info)
			return
		}

		if canInit, ok := op.(courier.CanInit); ok {
			if err := canInit.Init(ctx); err != nil {
				t.WriteResponse(ctx, rw, err, info)
				return
			}
		}

		result, err := op.Output(ctx)
		if err != nil {
			t.WriteResponse(ctx, rw, err, info)
			return
		}

		if !opFactory.IsLast {
			switch x := result.(type) {
			case courier.CanInjectContext:
				ctx = x.InjectContext(ctx)
			case context.Context:
				ctx = x
			default:
				if opFactory.ContextKey != nil {
					ctx = contextx.WithValue(ctx, opFactory.ContextKey, result)
				}
			}
			continue
		}

		t.WriteResponse(ctx, rw, result, info)
	}
}

type WithPreHandlerMiddleware interface {
	PreHandlerMiddleware(h http.Handler) http.Handler
}
