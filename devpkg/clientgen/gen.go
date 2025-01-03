package clientgen

import (
	"context"
	"errors"
	"fmt"
	"go/types"
	"log/slog"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/octohelm/courier/pkg/courierhttp/client"
	"github.com/octohelm/courier/pkg/openapi"
	"github.com/octohelm/courier/pkg/openapi/jsonschema"
	"github.com/octohelm/gengo/pkg/gengo"
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
	Decl   gengo.Snippet
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

	c.Render(gengo.Snippet{gengo.T: `
@doc
type @Operation struct {
	@courierhttpMethod@method ` + "`" + `path:@path` + "`" + `
	
	@Operation'Parameters
}

type @Operation'Parameters struct {
	@parameters
	@requestBody
}

@ResponseData
`,

		"contextContext":           gengo.ID("context.Context"),
		"courierhttpMethod":        gengo.ID("github.com/octohelm/courier/pkg/courierhttp.Method"),
		"courierMetadata":          gengo.ID("github.com/octohelm/courier/pkg/courier.Metadata"),
		"courierResult":            gengo.ID("github.com/octohelm/courier/pkg/courier.Result"),
		"courierClientFromContext": gengo.ID("github.com/octohelm/courier/pkg/courier.ClientFromContext"),
		"Operation":                gengo.ID(operation.OperationId),
		"method":                   gengo.ID(method),
		"path":                     path,
		"pkgName":                  c.Package("").Pkg().Name(),
		"doc":                      gengo.Comment(operation.Description),
		"ResponseData": func() gengo.Snippet {
			if hasResponse {
				return gengo.Snippet{gengo.T: `
func (r *@Operation) ResponseData() (*@Operation'Response) {
	return new(@Operation'Response)
}
`,
					"Operation": gengo.ID(operation.OperationId),
				}
			}

			return gengo.Snippet{gengo.T: `
func (r *@Operation) ResponseData() (*@courierNoContent) {
	return new(@courierNoContent)
}

`,
				"Operation":        gengo.ID(operation.OperationId),
				"courierNoContent": gengo.ID("github.com/octohelm/courier/pkg/courier.NoContent"),
			}
		}(),
		"parameters": gengo.MapSnippet(operation.Parameters, func(p *openapi.ParameterObject) gengo.Snippet {
			return gengo.Snippet{gengo.T: `
@doc
@FieldName @TypeDef ` + "`" + `name:@name in:@in@extraTag` + "`" + `
`,
				"name": func() any {
					if p.Required != nil && *p.Required {
						return p.Name
					}

					return fmt.Sprintf("%s,omitzero", p.Name)
				}(),
				"in": p.In,
				"FieldName": func() any {
					if goFieldName, ok := getSchemaExt(p.Schema, jsonschema.XGoFieldName); ok {
						return gengo.ID(goFieldName.(string))
					}
					return gengo.ID(gengo.UpperCamelCase(p.Name))
				}(),
				"TypeDef": func() any {
					s := g.typeOfSchema(c, p.Schema)

					if _, ok := getSchemaExt(p.Schema, jsonschema.XGoStarLevel); ok {
						return gengo.Snippet{gengo.T: `*@Type`, "Type": s}
					}

					return s
				}(),
				"extraTag": fieldPropExtraTag(p.Schema),
				"doc":      gengo.Comment(p.Description),
			}
		}),
		"requestBody": func(sw gengo.SnippetWriter) {
			if operation.RequestBody == nil || operation.RequestBody.Content == nil {
				return
			}

			for contentType := range operation.RequestBody.Content {
				mt := operation.RequestBody.Content[contentType]

				if contentType == "octet-stream" {
					sw.Render(gengo.Snippet{gengo.T: `
@Type ` + "`" + `in:"body" mime:@mime` + "`" + `
`,

						"mime": "application/octet-stream",
						"Type": gengo.ID("io.ReadCloser"),
					})
					continue
				}

				sw.Render(gengo.Snippet{gengo.T: `
@FieldName @Type ` + "`" + `in:"body" mime:@mime` + "`" + `
`,

					"mime": contentType,
					"Type": g.typeOfSchema(c, mt.Schema),
					"FieldName": func() any {
						if goFieldName, ok := getSchemaExt(mt.Schema, jsonschema.XGoFieldName); ok {
							return gengo.ID(goFieldName.(string))
						}
						return gengo.ID("")
					}(),
				})
			}
		},
	})

	return nil
}

func fieldPropExtraTag(s jsonschema.Schema) func(sw gengo.SnippetWriter) {
	return func(sw gengo.SnippetWriter) {
		if validate, ok := s.GetMetadata().GetExtension(jsonschema.XTagValidate); ok {
			sw.Render(gengo.Snippet{
				gengo.T:    " validate:@validate",
				"validate": validate.(string),
			})
		}

		if defaultValue := s.GetMetadata().Default; defaultValue != nil {
			sw.Render(gengo.Snippet{
				gengo.T:   " default:@default",
				"default": fmt.Sprintf("%v", defaultValue),
			})
		}
	}
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
				c.Render(gengo.Snippet{gengo.T: `
type @Type = @TypeRef
`,
					"Type":    gengo.ID(name),
					"TypeRef": gengo.ID(x),
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
		c.Render(gengo.Snippet{gengo.T: `
type @Type = @TypeDef

`,

			"Type":    gengo.ID(name),
			"TypeDef": t.Decl,
		})

		return nil
	}

	if unionType, ok := t.Schema.(*jsonschema.UnionType); ok {
		c.Render(gengo.Snippet{gengo.T: `
type @Type struct {
	Underlying any ` + "`" + `json:"-"` + "`" + `
}
`,
			"Type": gengo.ID(name),
		})

		if unionType.Discriminator != nil {
			c.Render(gengo.Snippet{gengo.T: `
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
`,
				"taggedunionUnmarshal": gengo.ID("github.com/octohelm/courier/pkg/validator/taggedunion.Unmarshal"),
				"jsonMarshal":          gengo.ID("github.com/octohelm/courier/pkg/validator.Marshal"),

				"Type":                      gengo.ID(name),
				"DiscriminatorPropertyName": unionType.Discriminator.PropertyName,
				"MappingValues": func(sw gengo.SnippetWriter) {
					for kind, s := range unionType.Discriminator.Mapping {

						sw.Render(gengo.Snippet{gengo.T: `
@Key: &@Type{},
`,
							"Key":  kind,
							"Type": g.typeOfSchema(c, s),
						})
					}
				},
			})
		}

		return nil
	}

	if enumType, ok := t.Schema.(*jsonschema.EnumType); ok {
		c.Render(gengo.Snippet{gengo.T: `
// +gengo:enum
type @Type @TypeDef

`,

			"Type":    gengo.ID(name),
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

		c.Render(gengo.Snippet{gengo.T: `
const (
	@enums
)

`,
			"enums": gengo.MapSnippet(enumType.Enum, func(enum any) gengo.Snippet {
				return gengo.Snippet{gengo.T: `
@NamePrefix'__@OrgName @Type = @value
`,
					"Type":       gengo.ID(gengo.UpperCamelCase(name)),
					"NamePrefix": gengo.ID(gengo.UpperSnakeCase(name)),
					"OrgName":    gengo.ID(gengo.UpperCamelCase(enum.(string))),
					"value":      enum,
				}
			}),
		})

		return nil
	}

	c.Render(gengo.Snippet{gengo.T: `
type @Type @TypeDef

`,

		"Type":    gengo.ID(name),
		"TypeDef": t.Decl,
	})

	return nil
}

func (g *clientGen) typeOfSchema(c gengo.Context, schema jsonschema.Schema) gengo.Snippet {
	switch x := schema.(type) {
	case *jsonschema.EnumType:
		return gengo.SnippetT("string")
	case *jsonschema.UnionType:
		// just look for walk sub schemas
		for _, s := range x.OneOf {
			g.typeOfSchema(c, s)
		}

		return gengo.SnippetT("struct { }")
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

			snippet := g.typeOfSchema(c, s)

			g.types.Store(name, &typ{
				Schema: s,
				Decl:   snippet,
			})
		}

		return gengo.Snippet{
			gengo.T: "@Type",
			"Type":  gengo.ID(name),
		}
	case *jsonschema.NumberType:
		if format, ok := x.GetExtension("x-format"); ok {
			return gengo.SnippetT(format.(string))
		}

		switch x.Type {
		case "integer":
			return gengo.SnippetT("int64")
		}
		return gengo.SnippetT("float64")
	case *jsonschema.StringType:
		switch x.Format {
		case "binary":
			return gengo.Snippet{
				gengo.T:        "@ioReadCloser",
				"ioReadCloser": gengo.ID("io.ReadCloser"),
			}
		}
		return gengo.SnippetT("string")
	case *jsonschema.BooleanType:
		return gengo.SnippetT("bool")
	case *jsonschema.ArrayType:
		if x.Items != nil {
			if x.MaxItems != nil && x.MinItems != nil && *x.MaxItems == *x.MinItems {
				return gengo.Snippet{gengo.T: "[@n]@TypeDef",
					"n":       *x.MaxItems,
					"TypeDef": g.typeOfSchema(c, x.Items),
				}
			}
		}

		return gengo.Snippet{gengo.T: "[]@TypeDef",
			"TypeDef": g.typeOfSchema(c, x.Items),
		}
	case *jsonschema.ObjectType:
		if elemSchema := x.AdditionalProperties; elemSchema != nil {
			var keySchema jsonschema.Schema = jsonschema.String()

			if x.PropertyNames != nil {
				keySchema = x.PropertyNames
			}

			return gengo.Snippet{gengo.T: `map[@KeyType]@ElemType`,
				"KeyType":  g.typeOfSchema(c, keySchema),
				"ElemType": g.typeOfSchema(c, elemSchema),
			}
		}

		return g.structFromSchema(c, x)
	}

	return gengo.SnippetT("any")
}

func (g *clientGen) structFromSchema(c gengo.Context, schema *jsonschema.ObjectType, extends ...jsonschema.Schema) gengo.Snippet {
	extendedDecls := make([]gengo.Snippet, len(extends))
	propDecls := map[string]gengo.Snippet{}

	for i := range extends {
		extendedDecls[i] = gengo.Snippet{gengo.T: `
@TypeDefEmbedded
`,
			"TypeDefEmbedded": g.typeOfSchema(c, extends[i]),
		}
	}

	requiredFieldSet := map[string]bool{}
	for _, name := range schema.Required {
		requiredFieldSet[name] = true
	}

	for name, propSchema := range schema.Properties.KeyValues() {
		propDecls[name] = gengo.Snippet{gengo.T: `
@doc
@FieldName @TypeDef ` + "`" + `json:@name name:@name @extraTags` + "`" + `
`,
			"name": func() any {
				if requiredFieldSet[name] {
					return name
				}
				return fmt.Sprintf("%s,omitempty", name)
			}(),
			"FieldName": func() any {
				if goFieldName, ok := getSchemaExt(propSchema, jsonschema.XGoFieldName); ok {
					return gengo.ID(goFieldName.(string))
				}
				return gengo.ID(gengo.UpperCamelCase(name))
			}(),
			"TypeDef": func() any {
				s := g.typeOfSchema(c, propSchema)

				if _, ok := getSchemaExt(propSchema, jsonschema.XGoStarLevel); ok {
					return gengo.Snippet{gengo.T: `*@Type`, "Type": s}
				}

				return s
			}(),
			"extraTags": fieldPropExtraTag(propSchema),
			"doc":       gengo.Comment(propSchema.GetMetadata().Description),
		}
	}

	return gengo.Snippet{gengo.T: `
struct { 
	@fields
} `,
		"fields": func(sw gengo.SnippetWriter) {
			for i := range extendedDecls {
				sw.Render(extendedDecls[i])
			}

			names := make([]string, 0)
			for fieldName := range propDecls {
				names = append(names, fieldName)
			}
			sort.Strings(names)

			for _, name := range names {
				sw.Render(propDecls[name])
			}
		},
	}
}

func getSchemaExt(schema jsonschema.Schema, name string) (any, bool) {
	if schema != nil {
		if v, ok := schema.GetMetadata().GetExtension(name); ok {
			return v, true
		}
	}

	return nil, false
}
