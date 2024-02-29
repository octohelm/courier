package openapi

import (
	"context"
	"fmt"
	"github.com/octohelm/courier/pkg/openapi/jsonschema"
	"github.com/octohelm/courier/pkg/ptr"
	"net/http"
	"reflect"
	"regexp"
	"strings"
	"sync"

	"github.com/octohelm/courier/internal/request"
	"github.com/octohelm/courier/pkg/courier"
	"github.com/octohelm/courier/pkg/courierhttp"
	"github.com/octohelm/courier/pkg/courierhttp/transport"
	"github.com/octohelm/courier/pkg/openapi"
	"github.com/octohelm/courier/pkg/openapi/jsonschema/extractors"
	"github.com/octohelm/courier/pkg/statuserror"
	transformer "github.com/octohelm/courier/pkg/transformer/core"
	"github.com/octohelm/gengo/pkg/gengo"
	typesx "github.com/octohelm/x/types"
	"github.com/pkg/errors"
)

type OpenAPIBuildFunc func(r courier.Router, fns ...BuildOptionFunc) *openapi.OpenAPI

var cached = sync.Map{}

var DefaultOpenAPIBuildFunc = func(r courier.Router, fns ...BuildOptionFunc) *openapi.OpenAPI {
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
		b.opt.naming = func(t string) string {
			splitter := map[string]bool{
				"internal": true,
				"pkg":      true,
				"apis":     true,
				"client":   true,
				"domain":   true,
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
				return gengo.UpperCamelCase(parts[0])
			}
			return gengo.UpperCamelCase(t)
		}
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
			mt := &openapi.MediaTypeObject{}
			mt.Schema = b.SchemaFromType(ctx, rt, false)
			resp.AddContent(contentType, mt)
		}
	} else {
		resp.AddContent(contentType, &openapi.MediaTypeObject{})
	}

	op.AddResponse(statusCode, resp)

	if can, ok := o.Operator.(CanResponseErrors); ok {
		returnErrors := can.ResponseErrors()

		codes := map[int][]string{}

		for i := range returnErrors {
			e := statuserror.FromErr(returnErrors[i])
			codes[e.StatusCode()] = append(codes[e.StatusCode()], e.Summary())
		}

		for statusCode := range codes {
			errResp := &openapi.ResponseObject{}
			mt := &openapi.MediaTypeObject{}
			mt.Schema = b.SchemaFromType(
				ctx,
				returnErrors[0],
				false,
			)

			errResp.AddContent("application/json", mt)
			errResp.AddExtension("x-status-return-errors", codes[statusCode])

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

	typesx.EachField(typesx.FromRType(t), "name", func(field typesx.StructField, fieldDisplayName string, omitempty bool) bool {
		location, _ := tagValueAndFlagsByTagString(field.Tag().Get("in"))

		if location == "" {
			panic(errors.Errorf("missing tag `in` for %s of %s", field.Name(), op.OperationId))
		}

		tf, err := transformer.NewTransformer(ctx, field.Type(), transformer.Option{
			MIME: strings.Split(field.Tag().Get("mime"), ",")[0],
		})
		if err != nil {
			panic(err)
		}

		v, _ := typesx.TryNew(field.Type())

		schema := b.SchemaFromType(ctx, v.Interface(), false)

		switch location {
		case "body":
			reqBody := op.RequestBody
			if op.RequestBody == nil {
				reqBody = &openapi.RequestBodyObject{
					Required: true,
				}
				op.SetRequestBody(reqBody)
			}

			if docer != nil {
				if schema != nil {
					if lines, ok := docer.RuntimeDoc(field.Name()); ok {
						schema.GetMetadata().Description = strings.Join(lines, "\n")
					}
				}
			}

			reqBody.AddContent(tf.Names()[0], &openapi.MediaTypeObject{
				Schema: schema,
			})
		case "query":
			op.AddParameter(fieldDisplayName, openapi.InQuery, &openapi.Parameter{
				Schema:   schema,
				Required: ptr.Ptr(!omitempty),
			})
		case "cookie":
			op.AddParameter(fieldDisplayName, openapi.InCookie, &openapi.Parameter{
				Schema:   schema,
				Required: ptr.Ptr(!omitempty),
			})
		case "header":
			op.AddParameter(fieldDisplayName, openapi.InHeader, &openapi.Parameter{
				Schema:   schema,
				Required: ptr.Ptr(!omitempty),
			})
		case "path":
			op.AddParameter(fieldDisplayName, openapi.InPath, &openapi.Parameter{
				Schema:   schema,
				Required: ptr.Ptr(true),
			})
		}

		return true
	}, "in")
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
