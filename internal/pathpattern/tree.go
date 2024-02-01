package pathpattern

import (
	"fmt"
	"io"
	"sort"
	"strings"
)

type Node interface {
	Method() string
	PathSegments() Segments
}

type Tree[N Node] struct {
	group[N]
}

func (t *Tree[N]) Add(n N) {
	t.group.add(n.Method(), n.PathSegments(), n)
}

func (t *Tree[N]) String() string {
	b := &strings.Builder{}
	t.group.PrintTo(b, 0)
	return b.String()
}

type group[N Node] struct {
	seg          Segment
	parent       *group[N]
	childExactly map[Segment]*group[N]
	childWild    *group[N]
	nodes        map[string]*N
}

type Route struct {
	Method        string
	PathSegments  Segments
	ChildSegments []Segment
}

func (r *Route) String() string {
	b := &strings.Builder{}
	b.WriteString(r.PathSegments.String())

	if len(r.ChildSegments) > 0 {
		b.WriteString("/(")
		for i, c := range r.ChildSegments {
			if i > 0 {
				b.WriteString("|")
			}
			b.WriteString(c.String())
		}
		b.WriteString(")")
	}

	return b.String()
}

func (g *group[N]) EachRoute(each func(n N, parents []*Route), parents ...*Route) {
	for _, node := range g.nodes {
		each(*node, parents)
	}

	var route *Route

	if named, ok := g.seg.(NamedSegment); ok && named.Multiple() {
		route = &Route{
			PathSegments: g.PathSegments(),
		}
	} else if len(g.childExactly) > 0 && g.childWild != nil {
		route = &Route{
			PathSegments: g.PathSegments(),
		}
	}

	if route != nil {
		for _, c := range g.childExactly {
			route.ChildSegments = append(route.ChildSegments, c.seg)
		}

		if c := g.childWild; c != nil {
			route.ChildSegments = append(route.ChildSegments, c.seg)
		}

		parents = append(parents, route)
	}

	for _, c := range g.childExactly {
		c.EachRoute(each, parents...)
	}

	if c := g.childWild; c != nil {
		c.EachRoute(each, parents...)
	}

}

func (g *group[N]) PathSegments() Segments {
	if g.parent != nil {
		return append(g.parent.PathSegments(), g.seg)
	}
	if g.seg != nil {
		return Segments{g.seg}
	}
	return Segments{}
}

func (g *group[N]) add(method string, segs Segments, node N) {
	if len(segs) == 0 {
		if g.nodes == nil {
			g.nodes = map[string]*N{}
		}
		g.nodes[method] = &node
		return
	}

	seg := segs[0]

	if named, ok := seg.(NamedSegment); ok {
		if currentNamed, ok := g.seg.(NamedSegment); ok && currentNamed.Multiple() {
			panic(fmt.Sprintf("named path segment is not allow after multiple segment: %s", node.PathSegments()))
		}

		if g.childWild != nil {
			if g.childWild.seg != named {
				panic(fmt.Sprintf("%s conflicts with %s", node.PathSegments(), g.childWild.PathSegments()))
			}
		} else {
			g.childWild = &group[N]{
				parent: g,
				seg:    named,
			}
		}

		g.childWild.add(method, segs[1:], node)
		return
	}

	if g.childExactly == nil {
		g.childExactly = map[Segment]*group[N]{}
	}

	child, ok := g.childExactly[seg]
	if !ok {
		child = &group[N]{
			seg:    seg,
			parent: g,
		}
		g.childExactly[seg] = child
	}

	child.add(method, segs[1:], node)
}

func (g *group[N]) PrintTo(w io.Writer, level int) {
	if g.seg != nil {
		if level > 0 {
			_, _ = fmt.Fprintf(w, strings.Repeat("  ", level))
		}

		_, _ = fmt.Fprintf(w, g.seg.String())
	}

	if len(g.nodes) > 0 {
		for _, node := range g.nodes {
			_, _ = fmt.Fprintf(w, "\n")
			if level > 0 {
				_, _ = fmt.Fprintf(w, strings.Repeat("  ", level+1))
			}
			_, _ = fmt.Fprintf(w, " => %s %s", (*node).Method(), (*node).PathSegments())
		}
		_, _ = fmt.Fprintf(w, "\n")
	} else {
		_, _ = fmt.Fprintf(w, "/")
		_, _ = fmt.Fprintf(w, "\n")
	}

	if g.childExactly != nil {
		exactlySegments := make([]Segment, 0, len(g.childExactly))
		for s := range g.childExactly {
			exactlySegments = append(exactlySegments, s)
		}
		sort.Slice(exactlySegments, func(i, j int) bool {
			return exactlySegments[i].String() < exactlySegments[j].String()
		})
		for _, seg := range exactlySegments {
			g.childExactly[seg].PrintTo(w, level+1)
		}
	}

	if g.childWild != nil {
		g.childWild.PrintTo(w, level+1)
	}
}
