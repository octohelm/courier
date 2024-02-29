package httprouter

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/fatih/color"
	"github.com/julienschmidt/httprouter"
	"github.com/octohelm/courier/internal/pathpattern"
	"github.com/octohelm/courier/internal/request"
	"github.com/octohelm/courier/pkg/courierhttp"
	"github.com/octohelm/courier/pkg/courierhttp/handler"
	openapispec "github.com/octohelm/courier/pkg/openapi"
	"github.com/pkg/errors"
)

type mux struct {
	oas              *openapispec.OpenAPI
	globalMiddleware handler.HandlerMiddleware
	w                *tabwriter.Writer
	tree             *pathpattern.Tree[RouteHandler]
}

func (m *mux) register(h request.RouteHandler) {
	if m.tree == nil {
		m.tree = &pathpattern.Tree[RouteHandler]{}
	}
	m.tree.Add(h)
}

func (m *mux) Handler() (http.Handler, error) {
	g := &group{}

	m.tree.EachRoute(func(h RouteHandler, parents []*pathpattern.Route) {
		if len(parents) == 0 {
			g.addHandler(h, []*pathpattern.Route{{
				PathSegments: h.PathSegments(),
			}})
		} else {
			g.addHandler(h, parents)
		}
	})

	_, _ = fmt.Fprintln(m.w)

	h, err := g.createHandler(m.w, func(ctx context.Context) context.Context {
		// FIXME only inject for openapi
		return openapispec.InjectContext(ctx, m.oas)
	})
	if err != nil {
		return nil, err
	}

	_, _ = fmt.Fprintln(m.w)

	return m.globalMiddleware(h), nil
}

type group struct {
	pathpattern.Route
	handlers []http.Handler
	children map[string]*group
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

func (g *group) addHandler(h RouteHandler, parents []*pathpattern.Route) {
	if g.children == nil {
		g.children = map[string]*group{}
	}

	if len(parents) == 0 {
		g.handlers = append(g.handlers, h)
		return
	}

	route := parents[0]

	child, ok := g.children[route.PathSegments.String()]
	if !ok {
		child = &group{Route: *route}
		g.children[route.PathSegments.String()] = child
	}

	child.addHandler(h, parents[1:])
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

func (g *group) createHandler(w *tabwriter.Writer, contextInjects ...contextInject) (h http.Handler, err error) {
	if len(g.handlers) > 0 {
		r := httprouter.New()

		for i := range g.handlers {
			h := g.handlers[i]

			if hh, ok := h.(RouteHandler); ok {
				ctxInjects := contextInjects[:]

				if rh, ok := h.(RouteHandler); ok {
					info := courierhttp.OperationInfo{
						Route:  rh.Path(),
						ID:     rh.OperationID(),
						Method: hh.Method(),
					}

					ctxInjects = append(ctxInjects, func(ctx context.Context) context.Context {
						return courierhttp.ContextWithOperationInfo(ctx, info)
					})
				}

				method := hh.Method()

				if method == "" {
					continue
				}

				pathSegments := hh.PathSegments()

				_, _ = colorForMethod(method).Fprintf(w, "%s\t  %s", method, pathSegments)
				_, _ = fmt.Fprintf(w, "\t    \t%s", hh.Summary())

				p := color.New(color.FgWhite)
				_, _ = p.Fprintf(w, "  {{ ")
				for i, o := range hh.Operators() {
					if i > 0 {
						_, _ = p.Fprint(w, " | ")
					}
					_, _ = p.Fprintf(w, "%s", o.String())
				}
				_, _ = p.Fprintf(w, " }}\n")

				r.Handle(method, toPath(pathSegments), func(rw http.ResponseWriter, req *http.Request, params httprouter.Params) {
					ctx := req.Context()
					for _, inject := range ctxInjects {
						ctx = inject(ctx)
					}
					ctx = handler.ContextWithParamGetter(ctx, params)
					hh.ServeHTTP(rw, req.WithContext(ctx))
				})
			} else {
				panic(errors.Errorf("invalid router %v", h))
			}
		}

		return r, nil
	}

	if n := len(g.children); n != 0 {
		r := httprouter.New()

		keys := make([]string, 0, n)
		for k := range g.children {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		for _, k := range keys {
			c := g.children[k]

			h, err := c.createHandler(w, contextInjects...)
			if err != nil {
				return nil, err
			}

			for _, m := range c.Methods() {
				if len(c.ChildSegments) > 0 {
					if c.PathMultiple() {
						prefix := c.PathSegments

						r.HandlerFunc(m, toPath(c.PathSegments), func(rw http.ResponseWriter, req *http.Request) {
							values := pathpattern.Values{}

							remain, ok := prefix.MatchTo(values, req.URL.Path)
							if ok {
								parts := strings.Split(remain, "/")

								for i, p := range parts {
									for _, seg := range c.ChildSegments {
										if p == seg.String() {
											multi := strings.Join(parts[0:i], "/")
											value := url.PathEscape(multi)

											r := req.Clone(req.Context())

											r.URL.Path = strings.Replace(req.URL.Path, multi, value, 1)
											r.RequestURI = r.URL.RequestURI()

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

func colorForMethod(method string) *color.Color {
	switch method {
	case http.MethodHead:
		return color.New(color.FgCyan)
	case http.MethodGet:
		return color.New(color.FgBlue)
	case http.MethodPost:
		return color.New(color.FgGreen)
	case http.MethodPut:
		return color.New(color.FgYellow)
	case http.MethodPatch:
		return color.New(color.FgMagenta)
	case http.MethodDelete:
		return color.New(color.FgRed)
	default:
		return color.New(color.FgWhite)
	}
}
