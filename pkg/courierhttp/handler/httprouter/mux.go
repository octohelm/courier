package httprouter

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"

	"github.com/juju/ansiterm"
	"github.com/julienschmidt/httprouter"
	"github.com/octohelm/courier/internal/pathpattern"
	"github.com/octohelm/courier/internal/request"
	"github.com/octohelm/courier/pkg/courierhttp"
	"github.com/octohelm/courier/pkg/courierhttp/handler"
	openapispec "github.com/octohelm/courier/pkg/openapi"
)

type mux struct {
	server courierhttp.Server
	oas    *openapispec.OpenAPI
	tree   *pathpattern.Tree[RouteHandler]
}

func (m *mux) register(h request.RouteHandler) {
	if m.tree == nil {
		m.tree = &pathpattern.Tree[RouteHandler]{}
	}
	m.tree.Add(h)
}

func (m *mux) Handler() (http.Handler, error) {
	g := &group{
		mux: m,
	}

	for h, parents := range m.tree.Route() {
		if len(parents) == 0 {
			g.addHandler(h, []*pathpattern.Route{{
				PathSegments: h.PathSegments(),
			}})
		} else {
			g.addHandler(h, parents)
		}
	}

	w := ansiterm.NewTabWriter(os.Stdout, 0, 4, 2, ' ', 0)
	defer func() {
		_ = w.Flush()
	}()

	_, _ = fmt.Fprintln(w)
	defer func() {
		_, _ = fmt.Fprintln(w)
	}()

	h, err := g.createHandler(w)
	if err != nil {
		return nil, err
	}

	return h, nil
}

type group struct {
	*mux
	parent *group
	pathpattern.Route
	handlers []RouteHandler
	children map[string]*group
}

func (g *group) String() string {
	b := bytes.NewBuffer(nil)
	g.debugTo(b, 0)
	return b.String()
}

func (g *group) debugTo(w io.Writer, level int) {
	_, _ = fmt.Fprint(w, strings.Repeat("....", level))
	_, _ = fmt.Fprintf(w, g.Route.PathSegments.String())
	_, _ = fmt.Fprintf(w, "\n")

	if len(g.children) > 0 {
		for _, child := range g.children {
			child.debugTo(w, level+1)
		}
	} else {
		for _, h := range g.handlers {
			if hh, ok := h.(RouteHandler); ok {
				_, _ = fmt.Fprint(w, strings.Repeat("....", level+1))
				_, _ = fmt.Fprintf(w, hh.PathSegments().String())
				_, _ = fmt.Fprintf(w, " ")
				_, _ = fmt.Fprintf(w, hh.Method())
				_, _ = fmt.Fprintf(w, "\n")
			}
		}
	}
}

func (g *group) PrintTo(w io.Writer, level int) {
	_, _ = fmt.Fprint(w, strings.Repeat("  ", level))
	_, _ = fmt.Fprintf(w, g.Route.PathSegments.String())
	_, _ = fmt.Fprintf(w, "\n")

	for _, child := range g.children {
		child.PrintTo(w, level+1)
	}

	for _, h := range g.handlers {
		if hh, ok := h.(RouteHandler); ok {
			_, _ = fmt.Fprint(w, strings.Repeat("  ", level+2))
			_, _ = fmt.Fprintf(w, hh.Method())
			_, _ = fmt.Fprintf(w, " ")
			_, _ = fmt.Fprintf(w, hh.PathSegments().String())
			_, _ = fmt.Fprintf(w, "\n")
		}
	}
}

func (g *group) child(route *pathpattern.Route) *group {
	path := route.PathSegments.String()

	child, ok := g.children[path]
	if ok {
		return child
	}

	child = &group{
		mux:    g.mux,
		parent: g,
		Route:  *route,
	}

	g.children[path] = child
	return child
}

func (g *group) addHandler(h RouteHandler, parents []*pathpattern.Route) {
	if len(parents) == 0 {
		if len(g.children) > 0 {
			child := g.child(&pathpattern.Route{
				PathSegments: h.PathSegments()[0 : len(g.PathSegments)+1],
			})
			child.addHandler(h, parents)
			return
		}

		g.handlers = append(g.handlers, h)
		return
	}

	if g.children == nil {
		g.children = map[string]*group{}
	}

	child := g.child(parents[0])

	if len(parents) > 0 {
		child.addHandler(h, parents[1:])
	}
}

func (g *group) Methods() []string {
	m := map[string]bool{}

	for _, h := range g.handlers {
		if rh, ok := h.(RouteHandler); ok {
			met := rh.Method()
			m[met] = true
		}
	}

	for _, child := range g.children {
		for _, method := range child.Methods() {
			m[method] = true
		}
	}

	methods := make([]string, 0, len(m))
	for met := range m {
		methods = append(methods, met)
	}
	sort.Strings(methods)

	return methods
}

type contextInject = func(ctx context.Context) context.Context

func (g *group) Path() string {
	return toPath(g.PathSegments)
}

func toPath(pathSegments pathpattern.Segments) string {
	s := &strings.Builder{}
	s.WriteString("/")

	segN := len(pathSegments)

	for i, seg := range pathSegments {
		if i > 0 {
			s.WriteString("/")
		}

		if named, ok := seg.(pathpattern.NamedSegment); ok {
			if named.Multiple() && i == (segN-1) {
				s.WriteString("*")
				s.WriteString(named.Name())
			} else {
				s.WriteString(":")
				s.WriteString(named.Name())
			}
		} else {
			s.WriteString(seg.String())
		}
	}

	return s.String()
}

func (g *group) newHandler(printer *ansiterm.TabWriter, handlers []RouteHandler, contextInjects ...contextInject) (h http.Handler, err error) {
	r := httprouter.New()

	for _, hh := range handlers {

		ctxInjects := contextInjects[:]
		info := courierhttp.OperationInfo{
			Server: g.mux.server,
		}

		info = courierhttp.OperationInfo{
			Server: g.mux.server,
			Route:  hh.Path(),
			Method: hh.Method(),
			ID:     hh.OperationID(),
		}

		ctxInjects = append(ctxInjects, func(ctx context.Context) context.Context {
			return courierhttp.ContextWithOperationInfo(ctx, info)
		})

		if info.Method == "GET" && info.ID == "OpenAPI" {
			ctxInjects = append(ctxInjects, func(ctx context.Context) context.Context {
				return openapispec.InjectContext(ctx, g.oas)
			})
		}

		method := hh.Method()

		if method == "" {
			continue
		}

		pathSegments := hh.PathSegments()

		colorFmt := colorFmtForMethod(method)

		_, _ = colorFmt.Fprint(printer, "%s", method)
		_, _ = colorFmt.Fprint(printer, "\t%s", pathSegments)
		_, _ = fmt.Fprintf(printer, "\t%s", hh.Summary())

		p := colorFormatter(ansiterm.Gray)
		_, _ = p.Fprint(printer, "\t{{ ")
		for i, o := range hh.Operators() {
			if i > 0 {
				_, _ = p.Fprint(printer, " | ")
			}
			_, _ = p.Fprint(printer, "%s", o.String())
		}
		_, _ = p.Fprint(printer, " }}\n")

		serverInfo := info.UserAgent()

		r.Handle(method, toPath(pathSegments), func(rw http.ResponseWriter, req *http.Request, params httprouter.Params) {
			ctx := req.Context()
			for _, inject := range ctxInjects {
				ctx = inject(ctx)
			}

			ctx = handler.ContextWithParamGetter(ctx, params)

			rw.Header().Set("Server", serverInfo)

			hh.ServeHTTP(rw, req.WithContext(ctx))
		})
	}

	return r, nil
}

func (g *group) createHandler(printer *ansiterm.TabWriter, contextInjects ...contextInject) (h http.Handler, err error) {
	if n := len(g.children); n != 0 {
		r := httprouter.New()

		keys := make([]string, 0, n)
		for k := range g.children {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		for _, k := range keys {
			c := g.children[k]

			h, err := c.createHandler(printer, contextInjects...)
			if err != nil {
				return nil, err
			}

			childSegments := c.AllChildSegments()

			for _, m := range c.Methods() {

				if len(childSegments) > 0 {
					if c.PathMultiple() {
						prefix := c.PathSegments

						r.HandlerFunc(m, toPath(c.PathSegments), func(rw http.ResponseWriter, req *http.Request) {
							values := pathpattern.Values{}

							remain, ok := prefix.MatchTo(values, req.URL.Path)
							if ok {
								parts := strings.Split(remain, "/")

								for i, p := range parts {
									for _, seg := range childSegments {
										if p == seg.String() {
											multi := strings.Join(parts[0:i], "/")
											value := url.PathEscape(multi)

											r := req.Clone(req.Context())

											u := *r.URL
											u.Path = strings.Replace(req.URL.Path, multi, value, 1)

											r.RequestURI = u.RequestURI()
											r.URL = &u

											h.ServeHTTP(rw, r)

											return
										}
									}
								}
							}

							rw.WriteHeader(http.StatusNotFound)
						})

						continue
					}

					r.Handle(m, toPath(c.PathSegments)+"/*path", func(rw http.ResponseWriter, req *http.Request, params httprouter.Params) {
						h.ServeHTTP(rw, req)
					})

					continue
				}

				r.Handle(m, toPath(c.PathSegments), func(rw http.ResponseWriter, req *http.Request, params httprouter.Params) {
					h.ServeHTTP(rw, req)
				})
			}
		}

		return r, nil
	}

	if len(g.handlers) > 0 {
		return g.newHandler(printer, g.handlers, contextInjects...)
	}

	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		rw.WriteHeader(http.StatusNotFound)
	}), nil
}

func (g *group) PathMultiple() bool {
	if len(g.PathSegments) > 0 {
		lastSeg := g.PathSegments[len(g.PathSegments)-1]
		if named, ok := lastSeg.(pathpattern.NamedSegment); ok {
			return named.Multiple()
		}
	}
	return false
}

func (g *group) AllChildSegments() (segments []pathpattern.Segment) {
	idx := len(g.PathSegments)

	if len(g.children) != 0 {
		for _, child := range g.children {
			segments = append(segments, child.PathSegments[idx])
		}
		return
	}

	for _, h := range g.handlers {
		if p := h.PathSegments(); idx < len(p) {
			segments = append(segments, p[idx])
		}
	}

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
