package extractors

import (
	"context"
	"encoding"
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"github.com/octohelm/courier/internal/jsonflags"
	"github.com/octohelm/courier/pkg/openapi/jsonschema"
	"github.com/octohelm/courier/pkg/validator"
	contextx "github.com/octohelm/x/context"
	reflectx "github.com/octohelm/x/reflect"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type RuntimeDocer interface {
	RuntimeDoc(names ...string) ([]string, bool)
}

var RuntimeDocerContext = contextx.New[RuntimeDocer]()

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

func must[T any](ret T, err error) T {
	if err != nil {
		panic(err)
	}
	return ret
}

func SchemaFromType(ctx context.Context, t reflect.Type, opt Opt) (s jsonschema.Schema) {
	sr := SchemaRegisterContext.From(ctx)

	fillGoVendorType := func(typeRef string) {
		if s != nil {
			if !(strings.Contains(typeRef, "/internal/") || strings.Contains(typeRef, "/internal.")) {
				s.GetMetadata().AddExtension(jsonschema.XGoVendorType, typeRef)
			}
		}
	}

	inst := reflect.New(t).Interface()

	// named type
	if pkgPath := t.PkgPath(); pkgPath != "" {
		typeRef := fmt.Sprintf("%s.%s", pkgPath, t.Name())

		defer fillGoVendorType(typeRef)

		ref := sr.RefString(typeRef)

		if ok := sr.Record(typeRef); ok {
			return &jsonschema.RefType{
				Ref: must(jsonschema.ParseURIReferenceString(ref)),
			}
		} else {
			defer func() {
				if n := len(opt.EnumInDoc); n > 0 {
					e := &jsonschema.EnumType{}

					e.Enum = make([]any, n)
					for i := range e.Enum {
						e.Enum[i] = opt.EnumInDoc[i]
					}

					if s != nil {
						s.GetMetadata().DeepCopyInto(e.GetMetadata())
					}

					s = e
				}

				sr.RegisterSchema(ref, s)

				if !opt.Decl {
					s = &jsonschema.RefType{Ref: must(jsonschema.ParseURIReferenceString(ref))}
				}
			}()
		}

		if canDoc, ok := inst.(jsonschema.CanSwaggerDoc); ok {
			opt = opt.WithDoc(canDoc.SwaggerDoc())
		}

		if canEnumValues, ok := inst.(jsonschema.GoEnumValues); ok {
			defer func() {
				values := canEnumValues.EnumValues()
				labels := make([]string, 0)
				e := &jsonschema.EnumType{}

				for i := range values {
					e.Enum = append(e.Enum, values[i])

					if canLabel, ok := values[i].(interface{ Label() string }); ok {
						labels = append(labels, canLabel.Label())
					}
				}

				if len(labels) > 0 {
					e.AddExtension(jsonschema.XEnumLabels, labels)
				}

				if s != nil {
					s.GetMetadata().DeepCopyInto(e.GetMetadata())
				}

				s = e
			}()
		}

		if docer, ok := inst.(RuntimeDocer); ok {
			ctx = RuntimeDocerContext.Inject(ctx, docer)

			defer func() {
				if lines, ok := docer.RuntimeDoc(); ok {
					SetTitleOrDescription(s.GetMetadata(), lines)
				}
			}()
		}

		defer fillGoVendorType(typeRef)

		if g, ok := inst.(jsonschema.GoUnionType); ok {
			if types := g.OneOf(); len(types) != 0 {
				schemas := make([]jsonschema.Schema, len(types))
				for i := range schemas {
					schemas[i] = SchemaFromType(ctx, reflectx.Deref(reflect.TypeOf(types[i])), opt.WithDecl(false))
				}
				if len(schemas) == 1 {
					return schemas[0]
				}
				return jsonschema.OneOf(schemas...)
			}
		}

		if g, ok := inst.(jsonschema.GoTaggedUnionType); ok {
			types := g.Mapping()

			tags := make([]string, 0, len(types))
			for tag := range types {
				tags = append(tags, tag)
			}
			sort.Strings(tags)

			schemas := make([]jsonschema.Schema, 0, len(types))
			mapping := map[string]jsonschema.Schema{}

			for _, tag := range tags {
				s := SchemaFromType(
					ctx,
					reflectx.Deref(reflect.TypeOf(types[tag])),
					opt.WithDecl(false),
				)

				schemas = append(schemas, s)
				mapping[tag] = s
			}

			s := jsonschema.OneOf(schemas...)

			s.Discriminator = &jsonschema.Discriminator{
				PropertyName: g.Discriminator(),
				Mapping:      mapping,
			}

			return s
		}

		if g, ok := inst.(jsonschema.OpenAPISchemaFormatGetter); ok {
			s := jsonschema.String()
			s.Format = g.OpenAPISchemaFormat()

			switch s.Format {
			case "int-or-string":
				return jsonschema.OneOf(jsonschema.Integer(), jsonschema.String())
			}

			return s
		}

		if g, ok := inst.(jsonschema.OpenAPISchemaTypeGetter); ok {
			typ := g.OpenAPISchemaType()
			if len(typ) > 0 && typ[0] != "" {
				p := jsonschema.Payload{}

				_ = p.UnmarshalJSON([]byte(fmt.Sprintf(`{"type":%q}`, typ[0])))

				if p.Schema != nil {
					return p.Schema
				}
			}
		}

		if g, ok := inst.(jsonschema.OpenAPISchemaGetter); ok {
			s := g.OpenAPISchema()
			return s
		}

		// TODO find better way
		if typeRef == "mime/multipart.FileHeader" || typeRef == "io.ReadCloser" {
			return jsonschema.Binary()
		}

		if _, ok := inst.(encoding.TextUnmarshaler); ok {
			if _, ok := inst.(encoding.TextMarshaler); ok {
				return jsonschema.String()
			}
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

		patch := func(s jsonschema.Schema) jsonschema.Schema {
			s.GetMetadata().AddExtension(jsonschema.XGoStarLevel, count)
			return s
		}

		return patch(s)
	case reflect.Interface:
		return jsonschema.Any()
	case reflect.String:
		return jsonschema.String()
	case reflect.Bool:
		return jsonschema.Boolean()
	case reflect.Float32:
		st := &jsonschema.NumberType{
			Type: "number",
		}
		st.AddExtension("x-format", "float32")
		return st
	case reflect.Float64:
		st := &jsonschema.NumberType{
			Type: "number",
		}
		st.AddExtension("x-format", "float64")
		return st
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		st := &jsonschema.NumberType{
			Type: "integer",
		}
		st.AddExtension("x-format", t.Kind().String())
		return st
	case reflect.Array:
		s := jsonschema.ArrayOf(SchemaFromType(ctx, t.Elem(), opt.WithDecl(false)))
		n := uint64(t.Len())
		s.MaxItems = &n
		s.MinItems = &n
		return s
	case reflect.Slice:
		if t.Elem().Kind() == reflect.Uint8 && t.Elem().PkgPath() == "" {
			return jsonschema.Bytes()
		}
		itemSchema := SchemaFromType(ctx, t.Elem(), opt.WithDecl(false))
		if itemSchema == nil {
			itemSchema = jsonschema.Any()
		}
		return jsonschema.ArrayOf(itemSchema)
	case reflect.Map:
		keySchema := SchemaFromType(ctx, t.Key(), opt.WithDecl(false))
		switch keySchema.(type) {
		case *jsonschema.StringType:
			break
		case *jsonschema.RefType:
			break
		default:
			if _, ok := keySchema.(*jsonschema.StringType); !ok {
				panic(fmt.Errorf("only support string of map key, but got %s", keySchema))
			}
		}
		return jsonschema.RecordOf(keySchema, SchemaFromType(ctx, t.Elem(), opt.WithDecl(false)))
	case reflect.Struct:
		structSchema := jsonschema.ObjectOf(nil)

		fields, err := jsonflags.Structs.StructFields(t)
		if err != nil {
			panic(err)
		}

		for f := range fields.StructField() {
			propSchema := toPropSchema(ctx, f, opt)

			if propSchema != nil {
				structSchema.SetProperty(f.Name, propSchema, !(f.Omitempty || f.Omitzero))
			}
		}

		v := reflect.New(t).Interface()
		if manifest, ok := v.(interface {
			GetObjectKind() schema.ObjectKind
		}); ok {
			apiVersion, kind := manifest.GetObjectKind().GroupVersionKind().ToAPIVersionAndKind()

			if kind != "" {
				if _, ok := structSchema.Properties.Get("kind"); ok {
					structSchema.SetProperty("kind", &jsonschema.EnumType{
						Enum: []any{kind},
					}, false)
				}
			}

			if apiVersion != "" {
				if _, ok := structSchema.Properties.Get("apiVersion"); ok {
					structSchema.SetProperty("apiVersion", &jsonschema.EnumType{
						Enum: []any{apiVersion},
					}, false)
				}
			}
		}

		return structSchema
	default:
		panic(fmt.Errorf("unsupported type %T", t))
	}

	return nil
}

func toPropSchema(ctx context.Context, sf *jsonflags.StructField, opt Opt) jsonschema.Schema {
	if !FieldShouldPick(sf.Type, sf.FieldName) {
		return nil
	}

	var fieldDoc []string

	if opt.Doc != nil {
		for _, name := range []string{
			sf.FieldName,
			sf.Name,
		} {
			if fieldDesc := opt.Doc[name]; fieldDesc != "" {
				stringEnum := pickStringEnumFromDesc(fieldDesc)
				if len(stringEnum) > 0 {
					opt = opt.WithEnumInDoc(stringEnum)
				}

				if i := strings.Index(fieldDesc, "."); i > 0 {
					fieldDoc = []string{fieldDesc[0:i], fieldDesc[i+1:]}
				} else if i := strings.Index(fieldDesc, "\n"); i > 0 {
					fieldDoc = []string{fieldDesc[0:i], fieldDesc[i+1:]}
				} else {
					fieldDoc = []string{fieldDesc, ""}
				}
			}
		}

	}

	propSchema := SchemaFromType(ctx, sf.Type, opt.WithDecl(false))
	if propSchema != nil {
		s, err := PatchSchemaValidation(propSchema, validator.Option{
			Type: sf.Type,
			Rule: sf.Tag.Get("validate"),
		})
		if err != nil {
			panic(fmt.Errorf("invalid validate %s: %w", sf.Tag.Get("validate"), err))
		}

		SetTitleOrDescription(s.GetMetadata(), fieldDoc)

		if canRuntimeDoc, ok := RuntimeDocerContext.MayFrom(ctx); ok {
			if lines, ok := canRuntimeDoc.RuntimeDoc(sf.FieldName); ok {
				SetTitleOrDescription(s.GetMetadata(), lines)
			}
		}

		s.GetMetadata().AddExtension(jsonschema.XGoFieldName, sf.FieldName)

		return s
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
