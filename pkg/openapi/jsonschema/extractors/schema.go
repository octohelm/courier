package extractors

import (
	"context"
	"fmt"
	"go/ast"
	"reflect"
	"strings"

	"github.com/octohelm/courier/pkg/openapi/jsonschema"
	"github.com/pkg/errors"
)

type EnumValues interface {
	EnumValues() []any
}

type RuntimeDocer interface {
	RuntimeDoc(names ...string) ([]string, bool)
}

type contextCanRuntimeDoc struct {
}

func ContextWithRuntimeDocer(ctx context.Context, sr RuntimeDocer) context.Context {
	return context.WithValue(ctx, contextCanRuntimeDoc{}, sr)
}

func RuntimeDocerFromContext(ctx context.Context) RuntimeDocer {
	if v, ok := ctx.Value(contextCanRuntimeDoc{}).(RuntimeDocer); ok {
		return v
	}
	return nil
}

type TypeName string

func (t TypeName) RefString() string {
	return string(t)
}

func SchemaFromType(ctx context.Context, t reflect.Type, def bool) (s *jsonschema.Schema) {
	sr := SchemaRegisterFromContext(ctx)

	// namedDefs typed
	if pkgPath := t.PkgPath(); pkgPath != "" {
		typeRef := fmt.Sprintf("%s.%s", pkgPath, t.Name())
		ref := sr.RefString(typeRef)

		if ok := sr.Record(typeRef); ok {
			if !def {
				return jsonschema.RefSchemaByRefer(TypeName(ref))
			}
		}

		v := reflect.New(t).Interface()

		if def {
			defer func() {
				if canEnumValues, ok := v.(EnumValues); ok {
					values := canEnumValues.EnumValues()
					labels := make([]string, 0)

					for i := range values {
						s.Enum = append(s.Enum, values[i])

						if canLabel, ok := values[i].(interface{ Label() string }); ok {
							labels = append(labels, canLabel.Label())
						}
					}

					if len(labels) > 0 {
						s.AddExtension(jsonschema.XEnumLabels, labels)
					}
				}
			}()

			if docer, ok := v.(RuntimeDocer); ok {
				ctx = ContextWithRuntimeDocer(ctx, docer)

				defer func() {
					if def {
						if lines, ok := docer.RuntimeDoc(); ok {
							s.Description = strings.Join(lines, "\n")
						}
					}
				}()
			}
		}

		defer func() {
			if !def {
				SchemaFromType(ctx, t, true)
			} else {
				sr.RegisterSchema(ref, s)
			}
		}()

		if g, ok := v.(OpenAPISchemaTypeGetter); ok {
			s := &jsonschema.Schema{}

			s.Type = g.OpenAPISchemaType()
			s.Format = ""

			if g, ok := v.(OpenAPISchemaFormatGetter); ok {
				s.Format = g.OpenAPISchemaFormat()
			}

			return s
		}

		if !def && t.Kind() != reflect.Interface {
			return jsonschema.RefSchemaByRefer(TypeName(ref))
		}

		defer func() {
			if s != nil {
				if !(strings.Contains(typeRef, "/internal/") || strings.Contains(typeRef, "/internal.")) {
					s.AddExtension(jsonschema.XGoVendorType, typeRef)
				}
			}
		}()

		for i := 0; i < t.NumMethod(); i++ {
			if t.Method(i).Name == "MarshalText" {
				return jsonschema.String()
			}
		}

		// TODO find better way
		if typeRef == "mime/multipart.FileHeader" {
			return jsonschema.Binary()
		}
	}

	switch t.Kind() {
	case reflect.Ptr:
		count := 1
		elem := t.Elem()

		for {
			if elem.Kind() == reflect.Ptr {
				elem = elem.Elem()
				count++
			} else {
				break
			}
		}

		s := SchemaFromType(ctx, elem, false)
		s.Nullable = true
		s.AddExtension(jsonschema.XGoStarLevel, count)

		return s
	case reflect.Interface:
		return &jsonschema.Schema{}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64,
		reflect.String,
		reflect.Invalid,
		reflect.Bool:
		return jsonschema.NewSchema(schemaTypeAndFormatFromBasicType(t.Kind().String()))
	case reflect.Array:
		s := jsonschema.ItemsOf(SchemaFromType(ctx, t.Elem(), false))
		n := uint64(t.Len())
		s.MaxItems = &n
		s.MinItems = &n
		return s
	case reflect.Slice:
		if t.Elem().Kind() == reflect.Uint8 && t.Elem().PkgPath() == "" {
			return jsonschema.Bytes()
		}

		return jsonschema.ItemsOf(SchemaFromType(ctx, t.Elem(), false))
	case reflect.Map:
		keySchema := SchemaFromType(ctx, t.Key(), false)
		if keySchema != nil && len(keySchema.Type) > 0 && !keySchema.Type.Contains("string") {
			panic(errors.New("only support map[string]any"))
		}
		return jsonschema.KeyValueOf(keySchema, SchemaFromType(ctx, t.Elem(), false))
	case reflect.Struct:
		structSchema := jsonschema.ObjectOf(nil)

		allOfSchemas := make([]*jsonschema.Schema, 0)

		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)

			if !ast.IsExported(field.Name) {
				continue
			}

			structTag := field.Tag

			tagValueForName := ""

			for _, namedTag := range []string{"json", "name"} {
				if tagValueForName == "" {
					tagValueForName = structTag.Get(namedTag)
				}
			}

			name, flags := tagValueAndFlagsByTagString(tagValueForName)
			if name == "-" {
				continue
			}

			if name == "" && field.Anonymous {
				if field.Type.String() == "bytes.Buffer" {
					structSchema = jsonschema.Binary()
					break
				}
				s := SchemaFromType(ctx, field.Type, false)
				if s != nil {
					allOfSchemas = append(allOfSchemas, s)
				}
				continue
			}

			if name == "" {
				name = field.Name
			}

			required := true

			if hasOmitempty, ok := flags["omitempty"]; ok {
				required = !hasOmitempty
			}

			propSchema := PropSchemaFromStructField(ctx, field, required)

			if propSchema != nil {
				structSchema.SetProperty(name, propSchema, required)
			}
		}

		if len(allOfSchemas) > 0 {
			return jsonschema.AllOf(append(allOfSchemas, structSchema)...)
		}

		return structSchema
	}

	return nil
}

func PropSchemaFromStructField(ctx context.Context, field reflect.StructField, required bool) *jsonschema.Schema {
	propSchema := SchemaFromType(ctx, field.Type, false)

	if propSchema != nil {
		if required {
			propSchema.Nullable = false
		}

		validate, hasValidate := field.Tag.Lookup("validate")

		if hasValidate && validate != "-" {
			if err := BindSchemaValidationByValidateBytes(propSchema, field.Type, []byte(validate)); err != nil {
				panic(errors.Wrapf(err, "invalid validate %s", validate))
			}
		}

		additional := &jsonschema.Schema{}

		if propSchema.Refer == nil {
			additional = propSchema
		}

		if canRuntimeDoc := RuntimeDocerFromContext(ctx); canRuntimeDoc != nil {
			if lines, ok := canRuntimeDoc.RuntimeDoc(field.Name); ok {
				additional.Description = strings.Join(lines, "\n")
			}
		}

		additional.AddExtension(jsonschema.XGoFieldName, field.Name)

		if propSchema != additional {
			return jsonschema.AllOf(propSchema, additional)
		}
		return propSchema
	}

	return nil
}

var basicTypeToSchemaType = map[string][2]string{
	"invalid": {"null", ""},

	"bool":    {"boolean", ""},
	"error":   {"string", "string"},
	"float32": {"number", "float"},
	"float64": {"number", "double"},

	"int":   {"integer", "int32"},
	"int8":  {"integer", "int8"},
	"int16": {"integer", "int16"},
	"int32": {"integer", "int32"},
	"int64": {"integer", "int64"},

	"rune": {"integer", "int32"},

	"uint":   {"integer", "uint32"},
	"uint8":  {"integer", "uint8"},
	"uint16": {"integer", "uint16"},
	"uint32": {"integer", "uint32"},
	"uint64": {"integer", "uint64"},

	"byte": {"integer", "uint8"},

	"string": {"string", ""},
}

func schemaTypeAndFormatFromBasicType(basicTypeName string) (typ string, format string) {
	if schemaTypeAndFormat, ok := basicTypeToSchemaType[basicTypeName]; ok {
		return schemaTypeAndFormat[0], schemaTypeAndFormat[1]
	}
	panic(errors.Errorf("unsupported type %q", basicTypeName))
}

func tagValueAndFlagsByTagString(tagString string) (string, map[string]bool) {
	valueAndFlags := strings.Split(tagString, ",")
	v := valueAndFlags[0]
	tagFlags := map[string]bool{}
	if len(valueAndFlags) > 1 {
		for _, flag := range valueAndFlags[1:] {
			tagFlags[flag] = true
		}
	}
	return v, tagFlags
}
