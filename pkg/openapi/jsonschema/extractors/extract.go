package extractors

import (
	"context"
	"reflect"
	"sync"

	"github.com/octohelm/courier/pkg/openapi/jsonschema"
)

func SchemaFrom(ctx context.Context, v any, def bool) jsonschema.Schema {
	if v == nil {
		return nil
	}

	t := reflect.TypeOf(v)

	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	return SchemaFromType(ctx, t, Opt{Decl: def})
}

type FieldExclude func(fields ...string)

type FieldFilter struct {
	Exclude []string
	Include []string
}

var fieldFilters sync.Map

func RegisterFieldFilter(t reflect.Type, fieldFilter FieldFilter) {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	fieldFilters.Store(t, fieldFilter)
}

func FieldShouldPick(t reflect.Type, fieldName string) bool {
	if filter, ok := fieldFilters.Load(t); ok {
		ff := filter.(FieldFilter)

		if (len(ff.Include)) > 0 {
			for _, include := range ff.Include {
				if include == fieldName {
					return true
				}
			}
			return false
		}

		if (len(ff.Exclude)) > 0 {
			for _, exclude := range ff.Exclude {
				if exclude == fieldName {
					return false
				}
			}
			return true
		}

		return false
	}

	return true
}
