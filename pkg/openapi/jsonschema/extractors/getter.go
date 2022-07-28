package extractors

import (
	"context"
	"reflect"

	"github.com/octohelm/courier/pkg/openapi/jsonschema"
)

func SchemaFrom(ctx context.Context, v any, def bool) (s *jsonschema.Schema) {
	if v == nil {
		return nil
	}

	defer func() {
		if !def {
			return
		}

		if g, ok := v.(OpenAPISchemaTypeGetter); ok {
			s.Type = g.OpenAPISchemaType()
			s.Format = ""
		}

		if g, ok := v.(OpenAPISchemaFormatGetter); ok {
			s.Format = g.OpenAPISchemaFormat()
		}
	}()

	t := reflect.TypeOf(v)

	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	return SchemaFromType(ctx, t, def)
}
