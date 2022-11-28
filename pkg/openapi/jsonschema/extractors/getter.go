package extractors

import (
	"context"
	"reflect"

	"github.com/octohelm/courier/pkg/openapi/jsonschema"
)

func SchemaFrom(ctx context.Context, v any, def bool) *jsonschema.Schema {
	if v == nil {
		return nil
	}

	t := reflect.TypeOf(v)

	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	return SchemaFromType(ctx, t, def)
}
