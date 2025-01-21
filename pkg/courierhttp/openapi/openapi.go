package openapi

import (
	"context"
	"fmt"
	"github.com/octohelm/courier/pkg/validator"
	"maps"
	"net/http"
	"reflect"
	"regexp"
	"slices"
	"strings"
	"sync"

	"github.com/octohelm/courier/internal/jsonflags"
	"github.com/octohelm/courier/internal/request"
	"github.com/octohelm/courier/pkg/content"
	"github.com/octohelm/courier/pkg/courier"
	"github.com/octohelm/courier/pkg/courierhttp"
	"github.com/octohelm/courier/pkg/courierhttp/transport"
	"github.com/octohelm/courier/pkg/openapi"
	"github.com/octohelm/courier/pkg/openapi/jsonschema"
	"github.com/octohelm/courier/pkg/openapi/jsonschema/extractors"
	"github.com/octohelm/courier/pkg/statuserror"
	"github.com/octohelm/gengo/pkg/gengo"
	"github.com/octohelm/x/ptr"
)

type BuildFunc func(r courier.Router, fns ...BuildOptionFunc) *openapi.OpenAPI

var cached = sync.Map{}

var DefaultBuildFunc = func(r courier.Router, fns ...BuildOptionFunc) *openapi.OpenAPI {
	if v, ok := cached.Load(r); ok {
		return v.(*openapi.OpenAPI)
	}
	o := FromRouter(r, fns...)
	cached.Store(r, o)
	return o
}

type CanResponseStatusCode interface {
	ResponseStatusCode() int
}

type CanResponseContentType interface {
	ResponseContentType() string
}

type CanResponseContent interface {
	ResponseContent() any
}

type CanResponseErrors interface {
	ResponseErrors() []error
}

func Naming(naming func(t string) string) BuildOptionFunc {
	return func(o *buildOption) {
		o.naming = naming
	}
}

var defaultPkgNamingPrefix = PkgNamingPrefix{}

func RegisterPkgNamingPrefix(pkgPath string, prefix string) {
	defaultPkgNamingPrefix.Register(pkgPath, prefix)
}

type PkgNamingPrefix map[string]string

func (p PkgNamingPrefix) Prefix(pkgPath string, name string) string {
	for _, pp := range slices.Sorted(maps.Keys(p)) {
		if strings.HasPrefix(pkgPath, pp) {
			return gengo.UpperCamelCase(p[pp] + "_" + name)
		}
	}

	return gengo.UpperCamelCase(name)
}

func (p PkgNamingPrefix) Register(pkgPath string, prefix string) {
	p[pkgPath] = prefix
}

type BuildOptionFunc func(o *buildOption)

type buildOption struct {
	naming func(t string) string
}

func FromRouter(r courier.Router, fns ...BuildOptionFunc) *openapi.OpenAPI {
	b := &scanner{
		o:   openapi.NewOpenAPI(),
		opt: buildOption{},
	}

	for i := range fns {
		fns[i](&b.opt)
	}

	if b.opt.naming == nil {
		naming := func(t string) string {
			var pkgPath string

			splitter := map[string]bool{
				"internal": true,
				"pkg":      true,
				"apis":     true,
				"api":      true,
				"client":   true,
				"domain":   true,
			}

			if i := strings.Index(t, "["); i > 0 {
				base := t[0:i]

				str := &strings.Builder{}

				for k, x := range strings.Split(t[i+1:len(t)-1], ",") {
					if k > 0 {
						str.WriteString("And")
					}
					str.WriteString(b.opt.naming(x))
				}

				str.WriteString("As")

				if j := strings.LastIndex(base, "."); j > 0 {
					pkgPath = base[0:j]
					str.WriteString(base[j+1:])
				} else {
					str.WriteString(base)
				}

				return defaultPkgNamingPrefix.Prefix(pkgPath, str.String())
			}

			if j := strings.LastIndex(t, "."); j > 0 {
				pkgPath = t[0:j]
			}

			parts := strings.Split(t, "/")

			idx := 0
			for i, p := range parts {
				if splitter[p] {
					idx = i
				}
			}

			if idx < len(parts)-1 {
				t = strings.Join(parts[idx+1:], "/")
			} else {
				t = strings.Join(parts[idx:], "/")
			}

			parts = strings.Split(t, ".")

			if len(parts) == 2 && strings.ToLower(parts[0]) == strings.ToLower(parts[1]) {
				return defaultPkgNamingPrefix.Prefix(pkgPath, parts[0])
			}

			return defaultPkgNamingPrefix.Prefix(pkgPath, t)
		}

		b.opt.naming = naming
	}

	routes := r.Routes()

	for i := range routes {
		if err := b.scan(routes[i]); err != nil {
			panic(err)
		}
	}

	return b.o
}

type scanner struct {
	o                 *openapi.OpenAPI
	m                 sync.Map
	incomingTransport transport.IncomingTransport
	opt               buildOption
}

func (b *scanner) Record(typeRef string) bool {
	_, ok := b.m.Load(typeRef)
	defer b.m.Store(typeRef, true)
	return ok
}

func tag(pkgPath string) string {
	tags := strings.Split(pkgPath, "/")
	return tags[len(tags)-1]
}

func (b *scanner) scan(r courier.Route) error {
	handlers, err := request.NewRouteHandlers(r, "openapi")
	if err != nil {
		return err
	}

	for _, rh := range handlers {
		op := openapi.NewOperation(rh.OperationID())

		op.Summary = rh.Summary()
		op.Description = rh.Description()

		if rh.Deprecated() {
			op.Deprecated = ptr.Ptr(true)
		}

		ctx := context.Background()

		for _, o := range rh.Operators() {
			b.scanParameterOrRequestBody(ctx, op, o.Type)

			if o.IsLast {
				/// response
				// FIXME make configurable
				op.Tags = []string{
					tag(o.Type.PkgPath()),
				}

				b.scanResponse(ctx, op, o)
			}

			b.scanResponseError(ctx, op, o)
		}

		b.o.AddOperation(rh.Method(), b.patchPath(rh.Path(), op), op)
	}

	return nil
}

var reHttpRouterPath = regexp.MustCompile("/{([^/]+)(...)?}")

func (b *scanner) patchPath(openapiPath string, operation *openapi.OperationObject) string {
	return reHttpRouterPath.ReplaceAllStringFunc(openapiPath, func(str string) string {
		name := reHttpRouterPath.FindAllStringSubmatch(str, -1)[0][1]

		if strings.HasSuffix(name, "...") {
			name = name[0 : len(name)-3]
		}

		isParameterDefined := false

		for _, parameter := range operation.Parameters {
			if parameter.In == "path" && parameter.Name == name {
				isParameterDefined = true
			}
		}

		if isParameterDefined {
			return "/{" + name + "}"
		}

		return "/0"
	})
}

func (b *scanner) RefString(ref string) string {
	return fmt.Sprintf("#/components/schemas/%s", b.opt.naming(ref))
}

func (b *scanner) RegisterSchema(ref string, s jsonschema.Schema) {
	if b.o.ComponentsObject.Schemas == nil {
		b.o.ComponentsObject.Schemas = map[string]jsonschema.Schema{}
	}

	n := strings.TrimLeft(ref, "#/components/schemas/")

	if _, ok := b.o.ComponentsObject.Schemas[n]; !ok {
		b.o.ComponentsObject.Schemas[n] = s
	} else {
		fmt.Println(n, "Registered.")
	}
}

func (b *scanner) SchemaFromType(ctx context.Context, v any, def bool) jsonschema.Schema {
	return extractors.SchemaFrom(extractors.SchemaRegisterContext.Inject(ctx, b), v, def)
}

func (b *scanner) scanResponse(ctx context.Context, op *openapi.OperationObject, o *courier.OperatorFactory) {
	method := ""

	statusCode := http.StatusNoContent
	contentType := "application/json"
	resp := &openapi.ResponseObject{}

	if can, ok := o.Operator.(courierhttp.MethodDescriber); ok {
		method = can.Method()

		if method == http.MethodPost {
			statusCode = http.StatusCreated
		} else {
			statusCode = http.StatusOK
		}
	}

	if method == "" {
		return
	}

	if can, ok := o.Operator.(CanResponseStatusCode); ok {
		statusCode = can.ResponseStatusCode()
	}

	if can, ok := o.Operator.(CanResponseContentType); ok {
		contentType = can.ResponseContentType()
	}

	if can, ok := o.Operator.(CanResponseContent); ok {
		if rt := can.ResponseContent(); rt != nil {
			if c, ok := rt.(courierhttp.ContentTypeDescriber); ok {
				contentType = c.ContentType()
			}

			mt := &openapi.MediaTypeObject{}
			mt.Schema = b.SchemaFromType(ctx, rt, false)
			resp.AddContent(contentType, mt)
		}
	} else {
		resp.AddContent(contentType, &openapi.MediaTypeObject{})
	}

	op.AddResponse(statusCode, resp)
}

func (b *scanner) scanResponseError(ctx context.Context, op *openapi.OperationObject, o *courier.OperatorFactory) {
	if can, ok := o.Operator.(CanResponseErrors); ok {
		returnErrors := can.ResponseErrors()

		codes := map[int][]string{}

		for _, err := range returnErrors {
			if e, ok := err.(statuserror.WithStatusCode); ok {
				if ok {
					codes[e.StatusCode()] = append(codes[e.StatusCode()], err.Error())
				}
			}
		}

		if op.Responses == nil {
			op.Responses = map[string]*openapi.ResponseObject{}
		}

		for statusCode := range codes {
			errResp, ok := op.Responses[fmt.Sprintf("%d", statusCode)]
			if !ok {
				errResp = &openapi.ResponseObject{}
			}

			mt := &openapi.MediaTypeObject{}

			switch x := returnErrors[0].(type) {
			case *statuserror.Descriptor:
				mt.Schema = b.SchemaFromType(
					ctx,
					&statuserror.ErrorResponse{},
					false,
				)
			default:
				mt.Schema = b.SchemaFromType(
					ctx,
					x,
					false,
				)
			}

			errResp.AddContent("application/json", mt)

			if found, ok := errResp.GetExtension("x-status-return-errors"); ok {
				errResp.AddExtension("x-status-return-errors", append(found.([]string), codes[statusCode]...))
			} else {
				errResp.AddExtension("x-status-return-errors", codes[statusCode])
			}

			op.AddResponse(statusCode, errResp)
		}
	}
}

type CanRuntimeDoc interface {
	RuntimeDoc(names ...string) ([]string, bool)
}

func (b *scanner) scanParameterOrRequestBody(ctx context.Context, op *openapi.OperationObject, t reflect.Type) {
	var docer CanRuntimeDoc

	if d, ok := reflect.New(t).Interface().(CanRuntimeDoc); ok {
		docer = d
	}

	fields, err := jsonflags.Structs.StructFields(t)
	if err != nil {
		panic(err)
	}

	for field := range fields.StructField() {
		location := field.Tag.Get("in")

		if location == "" {
			panic(fmt.Errorf("missing tag `in` for %s of %s", field.FieldName, op.OperationId))
		}
		optional := field.Omitzero || field.Omitempty

		tf, err := content.New(field.Type, field.Tag.Get("mime"), "unmarshal")
		if err != nil {
			panic(err)
		}

		schema := b.SchemaFromType(ctx, reflect.New(field.Type).Interface(), false)
		if schema != nil {
			_, err := extractors.PatchSchemaValidation(schema, validator.Option{
				Type: field.Type,
				Rule: field.Tag.Get("validate"),
			})

			if err != nil {
				panic(err)
			}
		}

		if schema != nil && docer != nil {
			if lines, ok := docer.RuntimeDoc(field.FieldName); ok {
				extractors.SetTitleOrDescription(schema.GetMetadata(), lines)
			}
		}

		switch location {
		case "body":
			reqBody := op.RequestBody
			if op.RequestBody == nil {
				reqBody = &openapi.RequestBodyObject{
					Required: true,
				}
				op.SetRequestBody(reqBody)
			}

			reqBody.AddContent(tf.MediaType(), &openapi.MediaTypeObject{
				Schema: schema,
			})
		case "query":
			op.AddParameter(field.Name, openapi.InQuery, &openapi.Parameter{
				Schema:   schema,
				Required: ptr.Ptr(!optional),
			})
		case "cookie":
			op.AddParameter(field.Name, openapi.InCookie, &openapi.Parameter{
				Schema:   schema,
				Required: ptr.Ptr(!optional),
			})
		case "header":
			op.AddParameter(field.Name, openapi.InHeader, &openapi.Parameter{
				Schema:   schema,
				Required: ptr.Ptr(!optional),
			})
		case "path":
			op.AddParameter(field.Name, openapi.InPath, &openapi.Parameter{
				Schema:   schema,
				Required: ptr.Ptr(true),
			})
		}
	}
}
