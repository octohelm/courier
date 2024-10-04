package httprouter

import (
	"bytes"
	"fmt"
	"github.com/octohelm/courier/internal/pathpattern"
	"iter"
	"maps"
	"net/http"
	"net/url"
	"slices"
	"strings"
)

type group struct {
	part     pathpattern.Segments
	parent   *group
	children map[string]*group
	handlers []RouteHandler
}

func (g *group) childSegment() iter.Seq[string] {
	if n := len(g.part); n > 0 {
		if named, ok := g.part[n-1].(pathpattern.NamedSegment); ok {
			if named.Multiple() {
				return func(yield func(string) bool) {
					emitted := map[string]bool{}

					for _, c := range g.children {
						if len(c.part) > 0 {
							seg := c.part[0].String()

							if emitted[seg] {
								continue
							}

							emitted[seg] = true

							if !yield(seg) {
								return
							}
						}
					}
				}

			}
		}
	}

	return func(yield func(string) bool) {

	}
}

func (g *group) handler(m *mux) http.Handler {
	if len(g.handlers) > 0 {
		r := http.NewServeMux()

		for _, h := range g.handlers {
			m.addHandler(r, h)
		}

		return r
	}

	r := http.NewServeMux()

	keys := slices.Sorted(maps.Keys(g.children))

	for _, k := range keys {
		c := g.children[k]

		hh := c.handler(m)
		prefix := c.pathSegments()
		childSegments := slices.Collect(c.childSegment())

		if len(childSegments) > 0 {
			r.HandleFunc(toHttpRouterPathPrefix(prefix), func(rw http.ResponseWriter, req *http.Request) {
				values := pathpattern.Values{}

				remain, ok := prefix.MatchTo(values, req.URL.Path)
				if ok {
					fmt.Println(remain)

					parts := strings.Split(remain, "/")

					for i, p := range parts {
						for _, seg := range childSegments {
							if p == seg {
								multi := strings.Join(parts[0:i], "/")
								value := url.PathEscape(multi)

								r := req.Clone(req.Context())

								u := *r.URL
								u.Path = strings.Replace(req.URL.Path, multi, value, 1)

								r.RequestURI = u.RequestURI()
								r.URL = &u

								hh.ServeHTTP(rw, r)

								return
							}
						}
					}
				}

				http.NotFound(rw, req)
			})

			continue
		}

		r.Handle(toHttpRouterPathPrefix(prefix), hh)
	}

	return r
}

func (g *group) String() string {
	b := bytes.NewBuffer(nil)

	d := g.depth()

	_, _ = fmt.Fprintf(b, "\n")
	_, _ = fmt.Fprintf(b, strings.Repeat("  ", d))
	_, _ = fmt.Fprintf(b, g.part.String())

	for _, c := range g.children {
		_, _ = fmt.Fprintf(b, c.String())
	}

	for _, h := range g.handlers {
		_, _ = fmt.Fprintf(b, "\n")
		_, _ = fmt.Fprintf(b, strings.Repeat("  ", d+1))
		_, _ = fmt.Fprintf(b, h.Method())
		_, _ = fmt.Fprintf(b, " ")
		_, _ = fmt.Fprintf(b, h.PathSegments().String())
	}

	return b.String()
}

func (g *group) depth() int {
	if g.parent == nil {
		return 0
	}
	return g.parent.depth() + 1
}

func (g *group) child(part pathpattern.Segments) *group {
	if g.children == nil {
		g.children = map[string]*group{}
	}

	p := part.String()

	child, ok := g.children[p]
	if !ok {
		c := &group{
			part:   part,
			parent: g,
		}

		g.children[p] = c

		return c
	}

	return child
}

func (g *group) add(h RouteHandler, chunk ...pathpattern.Segments) {
	if len(chunk) > 0 {
		g.child(chunk[0]).add(h, chunk[1:]...)
		return
	}
	g.handlers = append(g.handlers, h)
}

func (g *group) pathSegments() pathpattern.Segments {
	if p := g.parent; p != nil {
		return append(p.pathSegments(), g.part...)
	}
	return g.part
}
