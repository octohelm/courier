package clientgen

import (
	"context"
	"errors"
	"fmt"
	"go/types"
	"iter"
	"log/slog"
	"maps"
	"net/http"
	"net/url"
	"slices"
	"strconv"
	"strings"
	"sync"

	"github.com/octohelm/courier/pkg/courierhttp/client"
	"github.com/octohelm/courier/pkg/openapi"
	"github.com/octohelm/courier/pkg/openapi/jsonschema"
	"github.com/octohelm/gengo/pkg/gengo"
	"github.com/octohelm/gengo/pkg/gengo/snippet"
)

func init() {
	gengo.Register(&clientGen{})
}

type clientGen struct {
	types sync.Map
	oas   openapi.Payload
}

type typ struct {
	Alias  bool
	Schema jsonschema.Schema
	Decl   snippet.Snippet
}

func (g *clientGen) Name() string {
	return "client"
}

func (g *clientGen) GenerateType(c gengo.Context, named *types.Named) error {
	return g.generateClient(c, named)
}

func (g *clientGen) generateClient(c gengo.Context, named *types.Named) error {
	openapiSpec := ""
	tags, _ := c.Doc(named.Obj())

	includes := make([]string, 0)

	shouldGenerate := func(o *openapi.OperationObject) bool {
		if o == nil {
			return false
		}

		if len(includes) == 0 {
			return true
		}

		for i := range includes {
			if includes[i] == o.OperationId {
				return true
			}
		}

		return false
	}

	if r, ok := tags["gengo:client:openapi"]; ok {
		if len(r) > 0 {
			openapiSpec = r[0]
		}
	}
	if values, ok := tags["gengo:client:openapi:include"]; ok {
		includes = values
	}

	if openapiSpec == "" {
		return fmt.Errorf("openapi spec is not defined, please use `gengo:client:openapi=http://path/to/openapi/spec`")
	}

	u, err := url.Parse(openapiSpec)
	if err != nil {
		return err
	}

	switch u.Scheme {
	case "http", "https":
		cc := client.Client{
			Endpoint: u.Host,
		}
		req, _ := http.NewRequest("GET", u.String(), nil)
		_, err := cc.Do(context.Background(), req).Into(&g.oas)
		if err != nil {
			return errors.Join(gengo.ErrIgnore, err)
		}
	}

	for p, oo := range g.oas.Paths {
		for method, o := range oo.Operations {
			if shouldGenerate(o) {
				if err := g.genOperation(c, p, gengo.UpperCamelCase(strings.ToLower(method)), o); err != nil {
					return err
				}
			}

		}
	}

	var e error

	g.types.Range(func(k, value any) bool {
		t := value.(*typ)
		if err := g.genDef(c, k.(string), t); err != nil {
			e = err
			return false
		}
		return true
	})

	return e
}

func (g *clientGen) genOperation(c gengo.Context, path string, method string, operation *openapi.OperationObject) error {
	if operation.OperationId == "OpenAPI" {
		return nil
	}

	hasResponse := false
	for statusOrStr := range operation.ResponsesObject.Responses {
		status, _ := strconv.ParseInt(statusOrStr, 10, 64)

		if status >= http.StatusOK && status < http.StatusMultipleChoices {
			for _, mt := range operation.ResponsesObject.Responses[statusOrStr].Content {
				g.types.Store(fmt.Sprintf("%sResponse", operation.OperationId), &typ{
					Alias:  true,
					Schema: mt.Schema,
					Decl:   g.typeOfSchema(c, mt.Schema),
				})
				hasResponse = true
			}
		}
	}

	c.RenderT(`
@doc
type @Operation struct {
	@courierhttpMethod@method `+"`"+`path:@path`+"`"+`
	
	@Operation'Parameters
}

type @Operation'Parameters struct {
	@parameters
	@requestBody
}

@ResponseData
`, snippet.Args{
		"contextContext":           snippet.ID("context.Context"),
		"courierhttpMethod":        snippet.ID("github.com/octohelm/courier/pkg/courierhttp.Method"),
		"courierMetadata":          snippet.ID("github.com/octohelm/courier/pkg/courier.Metadata"),
		"courierResult":            snippet.ID("github.com/octohelm/courier/pkg/courier.Result"),
		"courierClientFromContext": snippet.ID("github.com/octohelm/courier/pkg/courier.ClientFromContext"),
		"Operation":                snippet.ID(operation.OperationId),
		"method":                   snippet.ID(method),
		"path":                     snippet.Value(path),
		"doc":                      snippet.Comment(operation.Description),
		"ResponseData": snippet.Snippets(func(yield func(snippet.Snippet) bool) {
			if hasResponse {
				if !yield(snippet.T(`
func (r *@Operation) ResponseData() (*@Operation'Response) {
	return new(@Operation'Response)
}
`, snippet.Args{
					"Operation": snippet.ID(operation.OperationId),
				})) {
					return
				}

				return
			}

			if !yield(snippet.T(`
func (r *@Operation) ResponseData() (*@courierNoContent) {
	return new(@courierNoContent)
}

`, snippet.Args{
				"Operation":        snippet.ID(operation.OperationId),
				"courierNoContent": snippet.ID("github.com/octohelm/courier/pkg/courier.NoContent"),
			})) {
				return
			}
			return
		}),
		"parameters": snippet.Snippets(func(yield func(snippet.Snippet) bool) {
			for _, p := range operation.Parameters {
				t := snippet.T(`
@doc
@FieldName @TypeDef `+"`"+`name:@name in:@in@extraTag`+"`"+`
`, snippet.Args{
					"name": snippet.Snippets(func(yield func(snippet.Snippet) bool) {
						if p.Required != nil && *p.Required {
							if !yield(snippet.Value(p.Name)) {
								return
							}
							return
						}
						if !yield(snippet.Value(fmt.Sprintf("%s,omitzero", p.Name))) {
							return
						}
						return
					}),
					"in": snippet.Value(p.In),
					"FieldName": snippet.Snippets(func(yield func(snippet.Snippet) bool) {
						if goFieldName, ok := getSchemaExt(p.Schema, jsonschema.XGoFieldName); ok {
							yield(snippet.ID(goFieldName.(string)))
							return
						}
						yield(snippet.ID(gengo.UpperCamelCase(p.Name)))
						return
					}),
					"TypeDef": snippet.Snippets(func(yield func(snippet.Snippet) bool) {
						s := g.typeOfSchema(c, p.Schema)

						if _, ok := getSchemaExt(p.Schema, jsonschema.XGoStarLevel); ok {
							if !yield(snippet.Sprintf(`*%T`, s)) {
								return
							}
							return
						}

						if !yield(snippet.Sprintf(`%T`, s)) {
							return
						}
					}),
					"extraTag": fieldPropExtraTag(p.Schema),
					"doc":      snippet.Comment(p.Description),
				})

				if !yield(t) {
					return
				}
			}
		}),
		"requestBody": snippet.Snippets(func(yield func(snippet.Snippet) bool) {
			if operation.RequestBody == nil || operation.RequestBody.Content == nil {
				return
			}

			for contentType := range operation.RequestBody.Content {
				mt := operation.RequestBody.Content[contentType]

				if contentType == "octet-stream" {
					if !yield(snippet.T(`
@Type `+"`"+`in:"body" mime:@mime`+"`"+`
`, snippet.Args{

						"mime": snippet.Value("application/octet-stream"),
						"Type": snippet.ID("io.ReadCloser"),
					})) {
						return
					}
					continue
				}

				if !yield(snippet.T(`
@FieldName @Type `+"`"+`in:"body" mime:@mime`+"`"+`
`, snippet.Args{

					"mime": snippet.Value(contentType),
					"Type": g.typeOfSchema(c, mt.Schema),
					"FieldName": func() snippet.Snippet {
						if goFieldName, ok := getSchemaExt(mt.Schema, jsonschema.XGoFieldName); ok {
							return snippet.ID(goFieldName.(string))
						}
						return snippet.ID("")
					}(),
				})) {
					return
				}
			}
		}),
	})

	return nil
}

func fieldPropExtraTag(s jsonschema.Schema) snippet.Snippet {
	return snippet.Func(func(ctx context.Context) iter.Seq[string] {
		return func(yield func(string) bool) {
			if validate, ok := s.GetMetadata().GetExtension(jsonschema.XTagValidate); ok {
				if !yield(fmt.Sprintf(" validate:%q", validate.(string))) {
					return
				}
			}

			if defaultValue := s.GetMetadata().Default; defaultValue != nil {
				if !yield(fmt.Sprintf(" default:%q", fmt.Sprintf("%v", defaultValue))) {
					return
				}
			}
		}
	})
}

func (g *clientGen) genDef(c gengo.Context, name string, t *typ) error {
	if name == "" {
		return fmt.Errorf("missing name of %s", t.Schema)
	}

	if t.Schema != nil {
		// when vendor imported in client, will be use the imported type
		if x, ok := t.Schema.GetMetadata().GetExtension(jsonschema.XGoVendorType); ok {
			imports := c.Package("").Imports()

			pkgPath, _ := gengo.PkgImportPathAndExpose(x.(string))

			if _, ok := imports[pkgPath]; ok {
				c.RenderT(`
type @Type = @TypeRef
`, snippet.Args{
					"Type":    snippet.ID(name),
					"TypeRef": snippet.ID(x),
				})

				return nil
			} else {
				if pkgPath != "" {
					c.Logger().WithValues(
						slog.String("import", pkgPath),
					).Info(fmt.Sprintf("not imported, will gen full type"))
				}
			}
		}
	}

	if t.Alias {
		c.RenderT(`
type @Type = @TypeDef

`, snippet.Args{

			"Type":    snippet.ID(name),
			"TypeDef": t.Decl,
		})

		return nil
	}

	if unionType, ok := t.Schema.(*jsonschema.UnionType); ok {
		c.RenderT(`
type @Type struct {
	Underlying any `+"`"+`json:"-"`+"`"+`
}
`, snippet.Args{
			"Type": snippet.ID(name),
		})

		if unionType.Discriminator != nil {
			c.RenderT(`
func (@Type) Discriminator() string {
	return @DiscriminatorPropertyName
}

func (@Type) Mapping() map[string]any {
	return map[string]any{
		@MappingValues
	}
}

func (m *@Type) SetUnderlying(v any) {
	m.Underlying = v
}

func (m *@Type) UnmarshalJSON(data []byte) error {
	mm := @Type{}
	if err := @taggedunionUnmarshal(data, &mm); err != nil {
		return err
	}
	*m = mm
	return nil
}

func (m @Type) MarshalJSON() ([]byte, error) {
	if m.Underlying == nil {
		return []byte("{}"), nil
	}
	return @jsonMarshal(m.Underlying)
}
`, snippet.Args{
				"taggedunionUnmarshal": snippet.ID("github.com/octohelm/courier/pkg/validator/taggedunion.Unmarshal"),
				"jsonMarshal":          snippet.ID("github.com/octohelm/courier/pkg/validator.Marshal"),

				"Type":                      snippet.ID(name),
				"DiscriminatorPropertyName": snippet.Value(unionType.Discriminator.PropertyName),
				"MappingValues": snippet.Snippets(func(yield func(snippet.Snippet) bool) {
					for _, kind := range slices.Sorted(maps.Keys(unionType.Discriminator.Mapping)) {
						s := unionType.Discriminator.Mapping[kind]

						if !yield(snippet.T(`
@Key: &@Type{},
`, snippet.Args{
							"Key":  snippet.Value(kind),
							"Type": g.typeOfSchema(c, s),
						})) {
							return
						}
					}
				}),
			})
		}

		return nil
	}

	if enumType, ok := t.Schema.(*jsonschema.EnumType); ok {
		c.RenderT(`
// +gengo:enum
type @Type @TypeDef

`, snippet.Args{
			"Type":    snippet.ID(name),
			"TypeDef": t.Decl,
		})

		enumLabels := make([]string, len(enumType.Enum))

		if xEnumLabels, ok := t.Schema.GetMetadata().GetExtension(jsonschema.XEnumLabels); ok {
			if labels, ok := xEnumLabels.([]interface{}); ok {
				for i, l := range labels {
					if v, ok := l.(string); ok {
						enumLabels[i] = v
					}
				}
			}
		}

		c.RenderT(`
const (
	@enums
)

`, snippet.Args{
			"enums": snippet.Snippets(func(yield func(snippet.Snippet) bool) {
				for _, enum := range enumType.Enum {
					if !yield(snippet.T(`
@NamePrefix'__@OrgName @Type = @value
`, snippet.Args{
						"Type":       snippet.ID(gengo.UpperCamelCase(name)),
						"NamePrefix": snippet.ID(gengo.UpperSnakeCase(name)),
						"OrgName":    snippet.ID(gengo.UpperCamelCase(enum.(string))),
						"value":      snippet.Value(enum),
					})) {
						return
					}
				}
			}),
		})

		return nil
	}

	c.RenderT(`
type @Type @TypeDef

`, snippet.Args{
		"Type":    snippet.ID(name),
		"TypeDef": t.Decl,
	})

	return nil
}

func (g *clientGen) typeOfSchema(c gengo.Context, schema jsonschema.Schema) snippet.Snippet {
	switch x := schema.(type) {
	case *jsonschema.EnumType:
		return snippet.Block("string")
	case *jsonschema.UnionType:
		// just look for walk sub schemas
		for _, s := range x.OneOf {
			g.typeOfSchema(c, s)
		}
		return snippet.Block("struct { }")
	case *jsonschema.IntersectionType:
		if len(x.AllOf) > 0 {
			// when one is the object
			if o, ok := x.AllOf[len(x.AllOf)-1].(*jsonschema.ObjectType); ok {
				return g.structFromSchema(c, o, x.AllOf[0:len(x.AllOf)-1]...)
			}
			return g.typeOfSchema(c, x)
		}
	case *jsonschema.RefType:
		name := x.Ref.RefName()

		if _, ok := g.types.Load(name); !ok {
			s := g.oas.Schemas[name]
			g.types.Store(name, nil)

			g.types.Store(name, &typ{
				Schema: s,
				Decl:   g.typeOfSchema(c, s),
			})
		}

		return snippet.ID(name)
	case *jsonschema.NumberType:
		if format, ok := x.GetExtension("x-format"); ok {
			return snippet.Block(format.(string))
		}

		switch x.Type {
		case "integer":
			return snippet.Block("int64")
		}
		return snippet.Block("float64")
	case *jsonschema.StringType:
		switch x.Format {
		case "binary":
			return snippet.ID("io.ReadCloser")
		}
		return snippet.Block("string")
	case *jsonschema.BooleanType:
		return snippet.Block("bool")
	case *jsonschema.ArrayType:
		if x.Items != nil {
			if x.MaxItems != nil && x.MinItems != nil && *x.MaxItems == *x.MinItems {
				return snippet.Sprintf("[%v]%T", *x.MaxItems, g.typeOfSchema(c, x.Items))
			}
		}
		return snippet.Sprintf("[]%T", g.typeOfSchema(c, x.Items))
	case *jsonschema.ObjectType:
		if elemSchema := x.AdditionalProperties; elemSchema != nil {
			var keySchema jsonschema.Schema = jsonschema.String()

			if x.PropertyNames != nil {
				keySchema = x.PropertyNames
			}

			return snippet.Sprintf("map[%T]%T", g.typeOfSchema(c, keySchema), g.typeOfSchema(c, elemSchema))
		}

		return g.structFromSchema(c, x)
	}

	return snippet.Block("any")
}

func (g *clientGen) structFromSchema(c gengo.Context, schema *jsonschema.ObjectType, extends ...jsonschema.Schema) snippet.Snippet {
	extendedDecls := make([]snippet.Snippet, len(extends))
	propDecls := map[string]snippet.Snippet{}

	for i := range extends {
		extendedDecls[i] = snippet.T(`
@TypeDefEmbedded
`, snippet.Args{
			"TypeDefEmbedded": g.typeOfSchema(c, extends[i]),
		})
	}

	requiredFieldSet := map[string]bool{}
	for _, name := range schema.Required {
		requiredFieldSet[name] = true
	}

	for name, propSchema := range schema.Properties.KeyValues() {
		propDecls[name] = snippet.T(`
@doc
@FieldName @TypeDef `+"`"+`json:@name name:@name @extraTags`+"`"+`
`, snippet.Args{
			"name": snippet.Value(func() any {
				if requiredFieldSet[name] {
					return name
				}
				return fmt.Sprintf("%s,omitempty", name)
			}()),
			"FieldName": func() snippet.Snippet {
				if goFieldName, ok := getSchemaExt(propSchema, jsonschema.XGoFieldName); ok {
					return snippet.ID(goFieldName.(string))
				}
				return snippet.ID(gengo.UpperCamelCase(name))
			}(),
			"TypeDef": func() snippet.Snippet {
				s := g.typeOfSchema(c, propSchema)

				if _, ok := getSchemaExt(propSchema, jsonschema.XGoStarLevel); ok {
					return snippet.Sprintf("*%T", s)
				}

				return s
			}(),
			"extraTags": fieldPropExtraTag(propSchema),
			"doc":       snippet.Comment(propSchema.GetMetadata().Description),
		})
	}

	return snippet.T(`
struct { 
	@fields
} `, snippet.Args{
		"fields": snippet.Snippets(func(yield func(snippet.Snippet) bool) {
			for _, decl := range extendedDecls {
				if !yield(decl) {
					return
				}
			}

			for _, name := range slices.Sorted(maps.Keys(propDecls)) {
				if !yield(propDecls[name]) {
					return
				}
			}
		}),
	})
}

func getSchemaExt(schema jsonschema.Schema, name string) (any, bool) {
	if schema != nil {
		if v, ok := schema.GetMetadata().GetExtension(name); ok {
			return v, true
		}
	}
	return nil, false
}
