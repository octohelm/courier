package pathpattern

import (
	"fmt"
	"io"
	"iter"
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
	childWild    *group[N]
	childExactly orderedMap[Segment, *group[N]]
	nodes        orderedMap[string, *N]
}

func (g *group[N]) Route() iter.Seq[N] {
	return func(yield func(N) bool) {
		for node := range g.nodes.Values() {
			if !(yield(*node)) {
				return
			}
		}

		for c := range g.childExactly.Values() {
			for node := range c.Route() {
				if !(yield(node)) {
					return
				}
			}
		}

		if c := g.childWild; c != nil {
			for node := range c.Route() {
				if !(yield(node)) {
					return
				}
			}
		}
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
		g.nodes.Add(method, &node)
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

	child, ok := g.childExactly.Get(seg)
	if !ok {
		child = &group[N]{
			seg:    seg,
			parent: g,
		}
		g.childExactly.Add(seg, child)
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

	if g.nodes.Len() > 0 {
		for node := range g.nodes.Values() {
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

	if n := g.childExactly.Len(); n > 0 {
		exactlySegments := make([]Segment, 0, n)
		for s := range g.childExactly.Keys() {
			exactlySegments = append(exactlySegments, s)
		}
		sort.Slice(exactlySegments, func(i, j int) bool {
			return exactlySegments[i].String() < exactlySegments[j].String()
		})

		for _, seg := range exactlySegments {
			v, _ := g.childExactly.Get(seg)
			v.PrintTo(w, level+1)
		}
	}

	if g.childWild != nil {
		g.childWild.PrintTo(w, level+1)
	}
}
