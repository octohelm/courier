package request

import (
	"context"
	"net/http"

	"github.com/octohelm/courier/internal/pathpattern"
	"github.com/octohelm/courier/pkg/courier"
	"github.com/octohelm/courier/pkg/courierhttp"
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

func NewRouteHandler(route courier.Route, service string) (RouteHandler, error) {
	h := &handler{
		service: service,
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

	if h.method == "" {
		h.method = "ALL"
	}

	return h, nil
}

type handler struct {
	service      string
	operationID  string
	method       string
	segments     pathpattern.Segments
	summary      string
	deprecated   bool
	description  string
	operators    []*courier.OperatorFactory
	transformers []transport.IncomingTransport
}

func (h *handler) OperationID() string {
	return h.operationID
}

func (h *handler) Method() string {
	return h.method
}

func (h *handler) Path() string {
	return h.segments.String()
}

func (h *handler) PathSegments() Segments {
	return h.segments
}

func (h *handler) Summary() string {
	return h.summary
}

func (h *handler) Description() string {
	return h.description
}

func (h *handler) Deprecated() bool {
	return h.deprecated
}

func (h *handler) Operators() []*courier.OperatorFactory {
	return h.operators
}

func (h *handler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	ctx = courierhttp.ContextWithHttpRequest(ctx, r)

	info := transport.FromHttpRequest(r, h.service)

	for i := range h.operators {
		opFactory := h.operators[i]
		t := h.transformers[i]

		op := opFactory.New()

		err := t.UnmarshalOperator(ctx, info, op)
		if err != nil {
			t.WriteResponse(ctx, rw, err, info)
			return
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
