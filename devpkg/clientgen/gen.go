package openapi

import (
	"context"
	"fmt"
	"go/types"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strings"

	"github.com/octohelm/courier/pkg/courierhttp/client"
	"github.com/octohelm/courier/pkg/openapi"
	"github.com/octohelm/courier/pkg/openapi/jsonschema"
	"github.com/octohelm/courier/pkg/transformer/core"
	"github.com/octohelm/gengo/pkg/gengo"
	"github.com/pkg/errors"
)

func init() {
	gengo.Register(&clientGen{})
}

type clientGen struct {
	schemas map[string]*openapi.Schema
	oas     openapi.OpenAPI
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

	if r, ok := tags["gengo:client:openapi"]; ok {
		if len(r) > 0 {
			openapiSpec = r[0]
		}
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
		cc := client.Client{Endpoint: u.Host}
		req, _ := http.NewRequest("GET", u.String(), nil)
		_, err := cc.Do(context.Background(), req).Into(&g.oas)
		if err != nil {
			return errors.Wrap(gengo.ErrIgnore, err.Error())
		}
	}

	for p, oo := range g.oas.Paths.Paths {
		for method, o := range oo.Operations.Operations {
			if err := g.genOperation(c, toColonPath(p), gengo.UpperCamelCase(strings.ToLower(string(method))), o); err != nil {
				return err
			}
		}
	}

	for name := range g.schemas {
		if err := g.genDef(c, name, g.schemas[name], false); err != nil {
			return err
		}
	}

	return nil
}

func (g *clientGen) genOperation(c gengo.Context, path string, method string, operation *openapi.Operation) error {
	if operation.OperationId == "OpenAPI" {
		return nil
	}

	hasResponse := false

	for status := range operation.Responses.Responses {
		if status >= http.StatusOK && status < http.StatusMultipleChoices {
			for _, mt := range operation.Responses.Responses[status].Content {
				if err := g.genDef(c, fmt.Sprintf("%sResponse", operation.OperationId), mt.Schema, true); err != nil {
					return err
				}

				hasResponse = true
			}
		}
	}

	c.Render(gengo.Snippet{gengo.T: `
@doc
type @Operation struct {
	@courierhttpMethod@method ` + "`" + `path:@path` + "`" + `
	@parameters
	@requestBody
}

func (r *@Operation) Do(ctx @contextContext, metas ...@courierMetadata) (@courierResult) {
	return @courierClientFromContent(ctx, @pkgName).Do(ctx, r, metas...)
}

@Invoke
`,

		"contextContext":           gengo.ID("context.Context"),
		"courierhttpMethod":        gengo.ID("github.com/octohelm/courier/pkg/courierhttp.Method"),
		"courierMetadata":          gengo.ID("github.com/octohelm/courier/pkg/courier.Metadata"),
		"courierResult":            gengo.ID("github.com/octohelm/courier/pkg/courier.Result"),
		"courierClientFromContent": gengo.ID("github.com/octohelm/courier/pkg/courier.ClientFromContent"),
		"Operation":                gengo.ID(operation.OperationId),
		"method":                   gengo.ID(method),
		"path":                     path,
		"pkgName":                  c.Package("").Pkg().Name(),
		"doc":                      gengo.Comment(operation.Description),
		"Invoke": func() gengo.Snippet {
			if hasResponse {
				return gengo.Snippet{gengo.T: `
func (r *@Operation) Invoke(ctx @contextContext, metas ...@courierMetadata) (*@Operation'Response, @courierMetadata, error) {
	var resp @Operation'Response
	meta, err := r.Do(ctx, metas...).Into(&resp)
	return &resp, meta, err
}

`,
					"Operation":       gengo.ID(operation.OperationId),
					"contextContext":  gengo.ID("context.Context"),
					"courierMetadata": gengo.ID("github.com/octohelm/courier/pkg/courier.Metadata"),
				}
			}
			return gengo.Snippet{gengo.T: `
func (r *@Operation) Invoke(ctx @contextContext, metas ...@courierMetadata) (@courierMetadata, error) {
	return r.Do(ctx, metas...).Into(nil)
}

`,
				"Operation":       gengo.ID(operation.OperationId),
				"contextContext":  gengo.ID("context.Context"),
				"courierMetadata": gengo.ID("github.com/octohelm/courier/pkg/courier.Metadata"),
			}
		},
		"parameters": gengo.MapSnippet(operation.Parameters, func(p *openapi.Parameter) gengo.Snippet {
			return gengo.Snippet{gengo.T: `
@doc
@FieldName @TypeDef ` + "`" + `name:@name in:@in@extraTag` + "`" + `
`,
				"name": func() any {
					if p.Required {
						return p.Name
					}
					return fmt.Sprintf("%s,omitempty", p.Name)
				}(),
				"in":        p.In,
				"FieldName": gengo.ID(gengo.UpperCamelCase(p.Name)),
				"TypeDef":   g.typeOfSchema(c, p.Schema),
				"extraTag":  fieldPropExtraTag(p.Schema),
				"doc":       gengo.Comment(p.Description),
			}
		}),
		"requestBody": func(sw gengo.SnippetWriter) {
			if operation.RequestBody == nil || operation.RequestBody.Content == nil {
				return
			}

			multi := len(operation.RequestBody.Content) > 1

			for contentType := range operation.RequestBody.Content {
				mt := operation.RequestBody.Content[contentType]

				mimeAlias := core.MgrFromContext(context.Background()).GetTransformerNames(contentType)[1]

				mime := func() any {
					if !multi {
						return mimeAlias
					}
					return fmt.Sprintf("%s,strict", mimeAlias)
				}()

				if mimeAlias == "octet-stream" {
					sw.Render(gengo.Snippet{gengo.T: `
@Type ` + "`" + `in:"body" mime:@mime` + "`" + `
`,

						"mime": mime,
						"Type": gengo.ID("io.ReadCloser"),
					})
					continue
				}

				if multi {
					sw.Render(gengo.Snippet{gengo.T: `
*@Type ` + "`" + `in:"body" mime:@mime` + "`" + `
`,

						"mime": mime,
						"Type": g.typeOfSchema(c, mt.Schema),
					})
					continue
				}

				sw.Render(gengo.Snippet{gengo.T: `
@Type ` + "`" + `in:"body" mime:@mime` + "`" + `
`,

					"mime": mime,
					"Type": g.typeOfSchema(c, mt.Schema),
				})
			}
		},
	})

	return nil
}

func fieldPropExtraTag(s *jsonschema.Schema) func(sw gengo.SnippetWriter) {
	return func(sw gengo.SnippetWriter) {
		if validate := s.Extensions[jsonschema.XTagValidate]; validate != nil {
			sw.Render(gengo.Snippet{
				gengo.T:    " validate:@validate",
				"validate": validate.(string),
			})
		}
	}
}

func (g *clientGen) genDef(c gengo.Context, name string, schema *jsonschema.Schema, alias bool) error {
	if schema != nil {
		if v := schema.Extensions[jsonschema.XGoVendorType]; v != nil {
			imports := c.Package("").Imports()
			pkgPath, expose := gengo.PkgImportPathAndExpose(v.(string))

			if _, ok := imports[pkgPath]; ok {

				c.Render(gengo.Snippet{gengo.T: `
type @Type = @TypeRef

`,
					"Type":    gengo.ID(gengo.UpperCamelCase(name)),
					"TypeRef": gengo.ID(pkgPath + "." + expose),
				})

				return nil
			}
		}
	}

	if alias {
		c.Render(gengo.Snippet{gengo.T: `
type @Type = @TypeDef

`,

			"Type":    gengo.ID(gengo.UpperCamelCase(name)),
			"TypeDef": g.typeOfSchema(c, schema),
		})

		return nil
	}

	c.Render(gengo.Snippet{gengo.T: `
type @Type @TypeDef

`,

		"Type":    gengo.ID(gengo.UpperCamelCase(name)),
		"TypeDef": g.typeOfSchema(c, schema),
	})

	if schema != nil && schema.Enum != nil {
		enumLabels := make([]string, len(schema.Enum))

		if xEnumLabels, ok := schema.Extensions[jsonschema.XEnumLabels]; ok {
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
			"enums": gengo.MapSnippet(schema.Enum, func(enum any) gengo.Snippet {
				return gengo.Snippet{gengo.T: `
@NamePrefix'__@Name @Type = @Value
`,
					"Type":       gengo.ID(gengo.UpperCamelCase(name)),
					"NamePrefix": gengo.ID(gengo.UpperSnakeCase(name)),
					"Name":       gengo.ID(gengo.UpperCamelCase(enum.(string))),
					"Value":      enum,
				}
			}),
		})
	}

	return nil
}

func (g *clientGen) typeOfSchema(c gengo.Context, schema *jsonschema.Schema) gengo.SnippetBuild {
	return func() gengo.Snippet {
		if schema == nil {
			return gengo.SnippetT("any")
		}

		if schema.Refer != nil {
			paths := schema.Refer.(*jsonschema.Ref).Paths

			name := paths[len(paths)-1]

			if g.schemas == nil {
				g.schemas = map[string]*openapi.Schema{}
			}

			g.schemas[name] = g.oas.Schemas[name]

			return gengo.Snippet{
				gengo.T: "@Type",
				"Type":  gengo.ID(gengo.UpperCamelCase(name)),
			}
		}

		if len(schema.AllOf) > 0 {
			// when one is the object
			if isObjectSchema(schema.AllOf[len(schema.AllOf)-1]) {
				return g.structFromSchema(c, schema.AllOf[len(schema.AllOf)-1], schema.AllOf[0:len(schema.AllOf)-1]...)()
			}
			return g.typeOfSchema(c, mayComposedAllOf(schema))()
		}

		typ := schema.Type[0]

		switch typ {
		case "object":
			if schema.AdditionalProperties != nil {
				keySchema := jsonschema.String()
				elemSchema := &jsonschema.Schema{}

				if schema.PropertyNames != nil {
					keySchema = schema.PropertyNames
				}

				if schema.AdditionalProperties.Schema != nil {
					elemSchema = schema.AdditionalProperties.Schema
				}

				return gengo.Snippet{gengo.T: `map[@KeyType]@ElemType`,
					"KeyType":  g.typeOfSchema(c, keySchema),
					"ElemType": g.typeOfSchema(c, elemSchema),
				}
			}

			return g.structFromSchema(c, schema)()
		case "array":
			if schema.Items != nil && schema.Items.Schema != nil {
				if schema.MaxItems != nil && schema.MinItems != nil && *schema.MaxItems == *schema.MinItems {
					return gengo.Snippet{gengo.T: "[@n]@TypeDef",
						"n":       *schema.MaxItems,
						"TypeDef": g.typeOfSchema(c, schema.Items.Schema),
					}
				}
			}

			return gengo.Snippet{gengo.T: "[]@TypeDef",
				"TypeDef": g.typeOfSchema(c, schema.Items.Schema),
			}

		default:
			return basicType(typ, schema.Format)()
		}
	}
}

func (g *clientGen) structFromSchema(c gengo.Context, schema *jsonschema.Schema, extends ...*jsonschema.Schema) func() gengo.Snippet {
	return func() gengo.Snippet {
		return gengo.Snippet{gengo.T: `
struct { 
	@fields
}
`,
			"fields": func(sw gengo.SnippetWriter) {
				for i := range extends {
					sw.Render(gengo.Snippet{gengo.T: `
@TypeDefEmbedded
`,
						"TypeDefEmbedded": g.typeOfSchema(c, extends[i]),
					})
				}

				names := make([]string, 0)
				for fieldName := range schema.Properties {
					names = append(names, fieldName)
				}
				sort.Strings(names)

				requiredFieldSet := map[string]bool{}
				for _, name := range schema.Required {
					requiredFieldSet[name] = true
				}

				for _, name := range names {
					propSchema := mayComposedAllOf(schema.Properties[name])

					sw.Render(gengo.Snippet{gengo.T: `
@doc
@FieldName @TypeDef ` + "`" + `json:@name name:@name @extraTags` + "`" + `
`,
						"name": func() any {
							if requiredFieldSet[name] {
								return name
							}
							return fmt.Sprintf("%s,omitempty", name)
						}(),
						"FieldName": gengo.ID(gengo.UpperCamelCase(name)),
						"TypeDef":   g.typeOfSchema(c, propSchema),
						"extraTags": fieldPropExtraTag(propSchema),
						"doc":       gengo.Comment(propSchema.Description),
					})
				}
			},
		}
	}
}

func isObjectSchema(schema *jsonschema.Schema) bool {
	t := schema.Type
	return len(t) > 0 && t[0] == jsonschema.TypeObject
}

func mayComposedAllOf(schema *jsonschema.Schema) *jsonschema.Schema {
	if schema.AllOf != nil && len(schema.AllOf) == 2 && !isObjectSchema(schema.AllOf[len(schema.AllOf)-1]) {
		nextSchema := &jsonschema.Schema{
			Reference:   schema.AllOf[0].Reference,
			SchemaBasic: schema.AllOf[1].SchemaBasic,
		}

		for k, v := range schema.AllOf[1].Extensions {
			nextSchema.AddExtension(k, v)
		}

		for k, v := range schema.Extensions {
			nextSchema.AddExtension(k, v)
		}

		return nextSchema
	}

	return schema
}

func basicType(schemaType string, format string) func() gengo.Snippet {
	return func() gengo.Snippet {
		switch format {
		case "binary":
			return gengo.Snippet{
				gengo.T:        "@ioReadCloser",
				"ioReadCloser": gengo.ID("io.ReadCloser"),
			}
		case "byte", "int", "int8", "int16", "int32", "int64", "rune", "uint", "uint8", "uint16", "uint32", "uint64", "uintptr", "float32", "float64":
			return gengo.SnippetT(format)
		case "float":
			return gengo.SnippetT("float32")
		case "double":
			return gengo.SnippetT("float64")
		default:
			switch schemaType {
			case "null":
				return gengo.SnippetT("any")
			case "integer":
				return gengo.SnippetT("int")
			case "number":
				return gengo.SnippetT("float64")
			case "boolean":
				return gengo.SnippetT("bool")
			default:
				return gengo.SnippetT("string")
			}
		}
	}
}

var reBraceToColon = regexp.MustCompile(`/\{([^/]+)\}`)

func toColonPath(path string) string {
	return reBraceToColon.ReplaceAllStringFunc(path, func(str string) string {
		name := reBraceToColon.FindAllStringSubmatch(str, -1)[0][1]
		return "/:" + name
	})
}
