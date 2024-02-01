package pathpattern

import (
	"fmt"
	"net/http"
	"testing"
)

func TestTree(t *testing.T) {
	tree := &Tree[*operation]{}

	tree.Add(Path(http.MethodGet, lit("v0"), lit("xxx")))
	tree.Add(Path(http.MethodPost, lit("v0"), lit("xxx")))
	tree.Add(Path(http.MethodGet, lit("v0"), lit("store"), namedMulti("scope"), lit("blobs"), lit("uploads")))
	tree.Add(Path(http.MethodGet, lit("v0"), lit("store"), namedMulti("scope"), lit("blobs"), named("digest")))
	tree.Add(Path(http.MethodGet, lit("v0"), lit("store"), namedMulti("scope"), lit("manifests"), named("reference")))

	tree.EachRoute(func(n *operation, parents []*Route) {
		fmt.Println(n.Method(), n.PathSegments(), parents)
	})
}

func Path(m string, segments ...Segment) *operation {
	return &operation{method: m, segments: segments}
}

type operation struct {
	method   string
	segments Segments
}

func (p operation) Method() string {
	return p.method
}

func (p operation) PathSegments() Segments {
	return p.segments
}
