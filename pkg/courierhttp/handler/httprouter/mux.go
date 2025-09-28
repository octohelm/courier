package httprouter

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"slices"

	"github.com/juju/ansiterm"
	"github.com/octohelm/courier/internal/pathpattern"
	"github.com/octohelm/courier/internal/request"
	"github.com/octohelm/courier/pkg/courierhttp"
	"github.com/octohelm/courier/pkg/courierhttp/handler"
)

type mux struct {
	server     courierhttp.Server
	operations *operations
	tree       *pathpattern.Tree[RouteHandler]
	w          *ansiterm.TabWriter
}

func (m *mux) register(h request.RouteHandler) {
	if m.tree == nil {
		m.tree = &pathpattern.Tree[RouteHandler]{}
	}
	m.tree.Add(h)
}

func (m *mux) Handler() (http.Handler, error) {
	w := ansiterm.NewTabWriter(os.Stdout, 0, 4, 2, ' ', 0)
	defer func() {
		_ = w.Flush()
	}()
	_, _ = fmt.Fprintln(w)
	defer func() {
		_, _ = fmt.Fprintln(w)
	}()
	m.w = w

	g := &group{}

	for h := range m.tree.Route() {
		g.add(h, slices.Collect(h.PathSegments().Chunk())...)
	}

	return g.handler(m), nil
}

type contextInject = func(ctx context.Context) context.Context

func (m *mux) addHandler(r *http.ServeMux, hh RouteHandler, contextInjects ...contextInject) {
	method := hh.Method()
	if method == "" {
		return
	}

	info := &courierhttp.OperationInfo{
		Server: m.server,
		Route:  hh.Path(),
		Method: hh.Method(),
		ID:     hh.OperationID(),
	}

	m.operations.add(info)

	ctxInjects := contextInjects[:]

	ctxInjects = append(ctxInjects, func(ctx context.Context) context.Context {
		return courierhttp.OperationInfoInjectContext(ctx, info)
	})

	ctxInjects = append(ctxInjects, func(ctx context.Context) context.Context {
		return courierhttp.OperationInfoProviderInjectContext(ctx, m.operations)
	})

	pathSegments := hh.PathSegments()

	colorFmt := colorFmtForMethod(method)

	_, _ = colorFmt.Fprint(m.w, "%s", method)
	_, _ = colorFmt.Fprint(m.w, "\t%s", pathSegments)
	_, _ = fmt.Fprintf(m.w, "\t%s", hh.Summary())

	p := colorFormatter(ansiterm.Gray)
	_, _ = p.Fprint(m.w, "\t{{ ")
	for i, o := range hh.Operators() {
		if i > 0 {
			_, _ = p.Fprint(m.w, " | ")
		}
		_, _ = p.Fprint(m.w, "%s", o.String())
	}
	_, _ = p.Fprint(m.w, " }}\n")

	serverInfo := info.UserAgent()

	r.HandleFunc(method+" "+toHttpRouterPathPrefix(pathSegments), func(rw http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		for _, inject := range ctxInjects {
			ctx = inject(ctx)
		}
		ctx = handler.ContextWithPathValueGetter(ctx, req)
		rw.Header().Set("Server", serverInfo)
		hh.ServeHTTP(rw, req.WithContext(ctx))
	})

	return
}

func colorFmtForMethod(method string) colorFormatter {
	switch method {
	case http.MethodHead:
		return colorFormatter(ansiterm.Cyan)
	case http.MethodGet:
		return colorFormatter(ansiterm.Blue)
	case http.MethodPost:
		return colorFormatter(ansiterm.Green)
	case http.MethodPut:
		return colorFormatter(ansiterm.Yellow)
	case http.MethodPatch:
		return colorFormatter(ansiterm.Magenta)
	case http.MethodDelete:
		return colorFormatter(ansiterm.Red)
	default:
		return colorFormatter(ansiterm.Gray)
	}
}

type colorFormatter ansiterm.Color

func (color colorFormatter) Fprint(w io.Writer, f string, args ...any) (int, error) {
	if x, ok := w.(interface {
		SetForeground(c ansiterm.Color)
	}); ok {
		x.SetForeground(ansiterm.Color(color))
		defer func() {
			x.SetForeground(ansiterm.Default)
		}()
	}
	return fmt.Fprintf(w, f, args...)
}
