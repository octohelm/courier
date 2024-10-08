package pathpattern

import (
	"fmt"
	"net/http"
	"slices"
	"testing"

	"github.com/davecgh/go-spew/spew"
)

func TestTree(t *testing.T) {
	tree := &Tree[*operation]{}

	tree.Add(createPath(http.MethodGet, lit("v0"), lit("xxx")))
	tree.Add(createPath(http.MethodPost, lit("v0"), lit("xxx")))
	tree.Add(createPath(http.MethodGet, lit("v0"), lit("store"), namedMulti("scope"), lit("blobs"), lit("uploads")))
	tree.Add(createPath(http.MethodGet, lit("v0"), lit("store"), namedMulti("scope"), lit("blobs"), named("digest")))
	tree.Add(createPath(http.MethodGet, lit("v0"), lit("store"), namedMulti("scope"), lit("manifests"), named("reference")))

	for n := range tree.Route() {
		fmt.Println(n.Method(), n.PathSegments())

		spew.Dump(slices.Collect(n.PathSegments().Chunk()))
	}
}

func createPath(m string, segments ...Segment) *operation {
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
