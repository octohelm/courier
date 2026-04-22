package pathpattern

import (
	"fmt"
	"net/http"
	"slices"
	"strings"
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

func TestTreeConflictErrorMessages(t *testing.T) {
	t.Run("多段通配后继续出现命名段应返回中文上下文", func(t *testing.T) {
		tree := &Tree[*operation]{}
		err := captureTreePanic(func() {
			tree.Add(createPath(http.MethodGet, lit("v0"), namedMulti("scope"), named("digest")))
		})
		if err == nil {
			t.Fatalf("expected panic error")
		}
		if !strings.Contains(err.Error(), "多段路径参数后不允许继续出现命名路径参数") {
			t.Fatalf("unexpected error message: %v", err)
		}
	})

	t.Run("命名路径冲突应返回中文上下文", func(t *testing.T) {
		tree := &Tree[*operation]{}
		tree.Add(createPath(http.MethodGet, lit("v0"), named("id")))
		err := captureTreePanic(func() {
			tree.Add(createPath(http.MethodGet, lit("v0"), named("name")))
		})
		if err == nil {
			t.Fatalf("expected panic error")
		}
		if !strings.Contains(err.Error(), "命名路径冲突") {
			t.Fatalf("unexpected error message: %v", err)
		}
		if !strings.Contains(err.Error(), "/v0/{name}") {
			t.Fatalf("missing conflicting path context: %v", err)
		}
	})
}

func captureTreePanic(fn func()) (err error) {
	defer func() {
		if x := recover(); x != nil {
			if e, ok := x.(error); ok {
				err = e
				return
			}
			err = fmt.Errorf("unexpected non-error panic: %v", x)
		}
	}()

	fn()
	return nil
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
