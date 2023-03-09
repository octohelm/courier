package extractors

import (
	"context"
	"fmt"
	reflectx "github.com/octohelm/x/reflect"
	"go/ast"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"github.com/octohelm/courier/pkg/openapi/jsonschema"
	"github.com/pkg/errors"
)

type EnumValues interface {
	EnumValues() []any
}

type UnionType interface {
	OneOf() []any
}

type TaggedUnionType interface {
	Discriminator() string
	Mapping() map[string]any
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

type Opt struct {
	Decl      bool
	Doc       map[string]string
	EnumInDoc []string
}

func (o Opt) WithDecl(decl bool) Opt {
	o.Decl = decl
	return o
}

func (o Opt) WithDoc(doc map[string]string) Opt {
	o.Doc = doc
	return o
}

func (o Opt) WithEnumInDoc(enumInDoc []string) Opt {
	o.EnumInDoc = enumInDoc
	return o
}

func SchemaFromType(ctx context.Context, t reflect.Type, opt Opt) (s *jsonschema.Schema) {
	sr := SchemaRegisterFromContext(ctx)

	// named type
	if pkgPath := t.PkgPath(); pkgPath != "" {
		typeRef := fmt.Sprintf("%s.%s", pkgPath, t.Name())
		ref := sr.RefString(typeRef)

		if ok := sr.Record(typeRef); ok {
			return jsonschema.RefSchemaByRefer(TypeName(ref))
		} else {
			defer func() {
				if n := len(opt.EnumInDoc); n > 0 {
					s.Enum = make([]any, n)
					for i := range s.Enum {
						s.Enum[i] = opt.EnumInDoc[i]
					}
				}

				sr.RegisterSchema(ref, s)

				if !opt.Decl {
					s = jsonschema.RefSchemaByRefer(TypeName(ref))
				}
			}()
		}

		inst := reflect.New(t).Interface()

		if canDoc, ok := inst.(CanSwaggerDoc); ok {
			opt = opt.WithDoc(canDoc.SwaggerDoc())
		}

		if canEnumValues, ok := inst.(EnumValues); ok {
			defer func() {
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
			}()
		}

		if docer, ok := inst.(RuntimeDocer); ok {
			ctx = ContextWithRuntimeDocer(ctx, docer)

			defer func() {
				if opt.Decl {
					if lines, ok := docer.RuntimeDoc(); ok {
						s.Description = strings.Join(lines, "\n")
					}
				}
			}()
		}

		if g, ok := inst.(UnionType); ok {
			types := g.OneOf()
			schemas := make([]*jsonschema.Schema, len(types))
			for i := range schemas {
				schemas[i] = SchemaFromType(ctx, reflectx.Deref(reflect.TypeOf(types[i])), opt.WithDecl(false))
			}
			return jsonschema.OneOf(schemas...)
		}

		if g, ok := inst.(TaggedUnionType); ok {
			types := g.Mapping()

			tags := make([]string, 0, len(types))
			for tag := range types {
				tags = append(tags, tag)
			}
			sort.Strings(tags)

			schemas := make([]*jsonschema.Schema, 0, len(types))
			for _, tag := range tags {
				schemas = append(schemas, SchemaFromType(ctx, reflectx.Deref(reflect.TypeOf(types[tag])), opt.WithDecl(true)))
			}
			s := jsonschema.OneOf(schemas...)

			s.Type = []string{jsonschema.TypeObject}
			s.Discriminator = &jsonschema.Discriminator{
				PropertyName: g.Discriminator(),
			}
			s.Required = []string{
				s.Discriminator.PropertyName,
			}

			return s
		}

		if g, ok := inst.(OpenAPISchemaTypeGetter); ok {
			s := &jsonschema.Schema{}

			s.Type = g.OpenAPISchemaType()
			s.Format = ""

			if g, ok := inst.(OpenAPISchemaFormatGetter); ok {
				s.Format = g.OpenAPISchemaFormat()
			}

			switch s.Format {
			case "int-or-string":
				return jsonschema.OneOf(jsonschema.Integer(), jsonschema.String())
			}

			return s
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

		s := SchemaFromType(ctx, elem, opt.WithDecl(false))

		patch := func(s *jsonschema.Schema) *jsonschema.Schema {
			s.Nullable = true
			s.AddExtension(jsonschema.XGoStarLevel, count)
			return s
		}

		if s.Refer != nil {
			return jsonschema.AllOf(s, patch(&jsonschema.Schema{}))
		}
		return patch(s)
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
		s := jsonschema.ItemsOf(SchemaFromType(ctx, t.Elem(), opt.WithDecl(false)))
		n := uint64(t.Len())
		s.MaxItems = &n
		s.MinItems = &n
		return s
	case reflect.Slice:
		if t.Elem().Kind() == reflect.Uint8 && t.Elem().PkgPath() == "" {
			return jsonschema.Bytes()
		}

		return jsonschema.ItemsOf(SchemaFromType(ctx, t.Elem(), opt.WithDecl(false)))
	case reflect.Map:
		keySchema := SchemaFromType(ctx, t.Key(), opt.WithDecl(false))
		if keySchema != nil && len(keySchema.Type) > 0 && !keySchema.Type.Contains("string") {
			panic(errors.New("only support map[string]any"))
		}
		return jsonschema.KeyValueOf(keySchema, SchemaFromType(ctx, t.Elem(), opt.WithDecl(false)))
	case reflect.Struct:
		structSchema := jsonschema.ObjectOf(nil)
		structSchema.AdditionalProperties = &jsonschema.SchemaOrBool{
			Allows: false,
		}

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

			// includes ,inline
			if name == "" && field.Anonymous {
				if field.Type.String() == "bytes.Buffer" {
					structSchema = jsonschema.Binary()
					break
				}
				s := SchemaFromType(ctx, field.Type, opt.WithDecl(false))
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

			propSchema := PropSchemaFromStructField(ctx, t, field, name, required, opt)

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

func PropSchemaFromStructField(
	ctx context.Context,
	t reflect.Type,
	field reflect.StructField,
	fieldName string,
	required bool,
	opt Opt,
) *jsonschema.Schema {
	if !FieldShouldPick(t, fieldName) {
		return nil
	}

	fieldDoc := ""

	if opt.Doc != nil {
		if fieldDesc := opt.Doc[fieldName]; fieldDesc != "" {
			fieldDoc = fieldDesc
			stringEnum := pickStringEnumFromDesc(fieldDesc)
			if len(stringEnum) > 0 {
				opt = opt.WithEnumInDoc(stringEnum)
			}
		}
	}

	propSchema := SchemaFromType(ctx, field.Type, opt.WithDecl(false))

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
		additional.Description = fieldDoc

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

func pickStringEnumFromDesc(d string) []string {
	parts := strings.Split(d, ".")
	for _, p := range parts {
		line := strings.TrimSpace(p)
		if strings.HasPrefix(line, "One of") {
			enumValues := strings.Split(line[len("One of")+1:], ",")
			for i := range enumValues {
				enumValues[i] = strings.TrimSpace(enumValues[i])
			}
			return enumValues
		}
		if strings.HasPrefix(line, "Can be") {
			enumValues := strings.Split(line[len("Can be")+1:], " or ")
			for i := range enumValues {
				enumValues[i] = strings.TrimSpace(enumValues[i])
				if len(enumValues[i]) > 0 {
					if enumValues[i][0] == '"' {
						enumValues[i], _ = strconv.Unquote(enumValues[i])
					}
				}
			}
			return enumValues
		}
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
