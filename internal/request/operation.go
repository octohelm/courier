package request

import (
	"reflect"
	"strings"

	"github.com/octohelm/courier/pkg/courier"
	"github.com/octohelm/courier/pkg/courierhttp"
)

type meta struct {
	OperationID string
	Method      string
	Path        string
	BasePath    string
	Summary     string
	Description string
	Deprecated  bool
}

var courierhttpPkgPath = reflect.TypeOf(courierhttp.MethodGet{}).PkgPath()

func metaFrom(o *courier.OperatorFactory) *meta {
	m := &meta{}

	op := o.Operator

	m.OperationID = o.Type.Name()

	if methodDescriber, ok := op.(courierhttp.MethodDescriber); ok {
		m.Method = methodDescriber.Method()
	}

	if canRuntimeDoc, ok := op.(CanRuntimeDoc); ok {
		if doc, ok := canRuntimeDoc.RuntimeDoc(); ok && len(doc) > 0 {
			m.Summary = doc[0]
			m.Description = strings.Join(doc[1:], "\n")
		}
	}

	if o.Type.Kind() == reflect.Struct {
		structType := o.Type

		for i := 0; i < structType.NumField(); i++ {
			f := structType.Field(i)
			if f.Anonymous && f.Type.PkgPath() == courierhttpPkgPath && strings.HasPrefix(f.Name, "Method") {
				if path, ok := f.Tag.Lookup("path"); ok {
					vs := strings.Split(path, ",")
					m.Path = vs[0]

					if len(vs) > 0 {
						for i := range vs {
							switch vs[i] {
							case "deprecated":
								m.Deprecated = true
							}
						}
					}
				}

				if basePath, ok := f.Tag.Lookup("basePath"); ok {
					m.BasePath = basePath
				}

				if summary, ok := f.Tag.Lookup("summary"); ok {
					m.Summary = summary
				}

				break
			}
		}
	}

	if basePathDescriber, ok := op.(courierhttp.BasePathDescriber); ok {
		m.BasePath = basePathDescriber.BasePath()
	}

	if pathDescriber, ok := op.(courierhttp.PathDescriber); ok {
		m.Path = pathDescriber.Path()
	}

	return m
}

type CanRuntimeDoc interface {
	RuntimeDoc(names ...string) ([]string, bool)
}
