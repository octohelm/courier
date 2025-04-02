package operatorgen

import (
	"fmt"
	"go/ast"
	"go/constant"
	"go/types"
	"reflect"
	"strconv"
	"strings"

	"github.com/octohelm/courier/pkg/courier"
	"github.com/octohelm/courier/pkg/courierhttp"
	"github.com/octohelm/courier/pkg/statuserror"
	"github.com/octohelm/gengo/pkg/gengo"
	"github.com/octohelm/gengo/pkg/gengo/snippet"
	gengotypes "github.com/octohelm/gengo/pkg/types"
	typex "github.com/octohelm/x/types"
)

func init() {
	gengo.Register(&operatorGen{})
}

type operatorGen struct{}

func (g *operatorGen) Name() string {
	return "operator"
}

var statusErrorScanner = newStatusErrScanner()

func (g *operatorGen) GenerateType(c gengo.Context, named *types.Named) error {
	if !ast.IsExported(named.Obj().Name()) {
		return gengo.ErrSkip
	}

	if !isCourierOperator(c, typex.FromTType(types.NewPointer(named)), g.resolvePkg) {
		return gengo.ErrSkip
	}

	g.generateRegister(c, named)
	g.generateReturns(c, named)
	return nil
}

func (g *operatorGen) generateReturns(c gengo.Context, named *types.Named) {
	method, ok := typex.FromTType(types.NewPointer(named)).MethodByName("Output")
	if ok {
		results, n := c.Package(named.Obj().Pkg().Path()).ResultsOf(method.(*typex.TMethod).Func)
		if n == 2 {
			g.generateSuccessReturn(c, named, results[0])
			g.generateErrorsReturn(c, named, method.(*typex.TMethod).Func)
		}
	}
}

func (g *operatorGen) generateErrorsReturn(c gengo.Context, named *types.Named, fn *types.Func) {
	statusErrors := statusErrorScanner.StatusErrorsInFunc(c, fn)
	if len(statusErrors) > 0 {
		c.RenderT(`
func (@Type) ResponseErrors() []error {
	return []error{
		@statusErrors
	}
}

`, snippet.Args{
			"Type": snippet.ID(named.Obj()),
			"statusErrors": snippet.Snippets(func(yield func(snippet.Snippet) bool) {
				added := map[string]bool{}

				for _, statusError := range statusErrors {
					x := statusError.Error()
					if _, ok := added[x]; ok {
						continue
					}

					if !strings.Contains(statusError.Code, "/") {
						if !yield(snippet.Sprintf("%v,\n", statusError)) {
							return
						}
					} else {
						if !yield(snippet.T(`
&@statuserrorDescriptor{
	Code:    @statuserrorErrCodeFor[@Error](),
	Message: @message,
	Status:  @status,
},
`, snippet.Args{
							"Error":                 snippet.ID(statusError.Code),
							"statuserrorDescriptor": snippet.PkgExposeFor[statuserror.Descriptor](),
							"statuserrorErrCodeFor": snippet.PkgExposeFor[statuserror.Descriptor]("ErrCodeFor"),
							"message":               snippet.Value(statusError.Message),
							"status":                snippet.Value(statusError.Status),
						})) {
							return
						}
					}

					added[x] = true
				}
			}),
		})
	}
}

func (g *operatorGen) generateSuccessReturn(c gengo.Context, named *types.Named, typeAndValues gengotypes.TypeAndValues) {
	var successType types.Type
	var fromExpr ast.Expr

	for _, resp := range typeAndValues {
		if resp.Type != nil {
			tpe := dePtr(resp.Type)

			if isNil(tpe) {
				continue
			}

			typeStr := tpe.String()

			if strings.HasSuffix(typeStr, "/courierhttp.Response[any]") {
				continue
			}

			if !isNil(successType) {
				if successType.String() != typeStr {
					c.Logger().Warn(fmt.Errorf("%s return multi types, `%s` `%s`", named, successType, tpe))
				}
			}

			successType = tpe
			fromExpr = resp.Expr
			// use first got type
			break
		}
	}

	if isNil(successType) {
		c.RenderT(`
func (@Type) ResponseContent() any {
	return nil
}

func (@Type) ResponseData() *@courierNoContent {
	return new(@courierNoContent)
}

`, snippet.Args{
			"Type":             snippet.ID(named.Obj()),
			"courierNoContent": snippet.ID("github.com/octohelm/courier/pkg/courier.NoContent"),
		})
	} else if types.IsInterface(successType) && !strings.Contains(successType.String(), "github.com/octohelm/courier/pkg/courierhttp.Response") {
		c.Logger().Warn(fmt.Errorf("%s return interface %s will be untyped jsonschema", named, successType))
	} else {
		if n, ok := successType.(*types.Named); ok {
			typeArgs := n.TypeArgs()

			if typeArgs.Len() > 0 {
				if n.Obj().Pkg().Path() == typeResponse.PkgPath() && n.Obj().Name() == "Response" {
					successType = dePtr(n.TypeArgs().At(0))

					if fromExpr != nil {
						ast.Inspect(fromExpr, func(node ast.Node) bool {
							switch callExpr := node.(type) {
							case *ast.CallExpr:
								switch e := callExpr.Fun.(type) {
								case *ast.SelectorExpr:
									switch e.Sel.Name {
									case "WithStatusCode", "Redirect":
										if p := c.LocateInPackage(node.Pos()); p != nil {
											v, err := p.Eval(callExpr.Args[0])
											if err != nil {
												return true
											}

											if statueCode, ok := valueOf(v.Value).(int64); ok {
												c.RenderT(`
func (@Type) ResponseStatusCode() int {
	return @statueCode
}

`, snippet.Args{
													"Type":       snippet.ID(named.Obj()),
													"statueCode": snippet.Value(int(statueCode)),
												})
											}
										}
										return false
									case "WithContentType":
										if p := c.LocateInPackage(node.Pos()); p != nil {
											v, _ := p.Eval(callExpr.Args[0])
											if contentType, ok := valueOf(v.Value).(string); ok {
												c.RenderT(`
func (@Type) ResponseContentType() string {
	return @contentType
}

`, snippet.Args{
													"Type":        snippet.ID(named.Obj()),
													"contentType": snippet.Value(contentType),
												})
											}
										}
										return false
									}
								}
							}
							return true
						})
					}
				}
			}
		}

		if _, ok := successType.(*types.Interface); ok {
			return
		}

		c.RenderT(`
func (@Type) ResponseContent() any {
	return new(@ReturnType)
}

func (@Type) ResponseData() *@ReturnType {
	return new(@ReturnType)
}

`, snippet.Args{
			"Type":       snippet.ID(named.Obj()),
			"ReturnType": snippet.ID(successType),
		})
	}
}

func dePtr(t types.Type) types.Type {
	if p, ok := t.(*types.Pointer); ok {
		t = p.Elem()
	}
	return t
}

var typeResponse = reflect.TypeFor[courierhttp.Response[any]]()

func (g *operatorGen) generateRegister(c gengo.Context, named *types.Named) {
	tags, _ := c.Doc(named.Obj())

	if registers, ok := tags["gengo:operator:register"]; ok {
		for _, register := range registers {
			c.RenderT(`
			
func init() {
	@R.Register(@courierNewRouter(&@Operator{}))
}
			
			`, snippet.Args{
				"R":                snippet.ID(register),
				"courierNewRouter": snippet.ID("github.com/octohelm/courier/pkg/courier.NewRouter"),
				"Operator":         snippet.ID(named.Obj()),
			})
		}
	}
}

func (g *operatorGen) resolvePkg(c gengo.Context, importPath string) *types.Package {
	return c.Package(importPath).Pkg()
}

func (g *operatorGen) firstValueOfFunc(c gengo.Context, named *types.Named, name string) (any, bool) {
	method, ok := typex.FromTType(types.NewPointer(named)).MethodByName(name)
	if ok {
		fn := method.(*typex.TMethod).Func
		results, n := c.Package(fn.Pkg().Path()).ResultsOf(fn)
		if n == 1 {
			for _, r := range results[0] {
				if v := valueOf(r.Value); v != nil {
					return v, true
				}
			}
			return nil, true
		}
	}
	return nil, false
}

var typOperator = reflect.TypeOf((*courier.Operator)(nil)).Elem()

func isCourierOperator(c gengo.Context, tpe typex.Type, lookup func(c gengo.Context, importPath string) *types.Package) bool {
	switch tpe.(type) {
	case *typex.RType:
		return tpe.Implements(typex.FromRType(typOperator))
	case *typex.TType:
		pkg := lookup(c, typOperator.PkgPath())
		if pkg == nil {
			return false
		}
		t := pkg.Scope().Lookup(typOperator.Name())
		if t == nil {
			return false
		}
		return types.Implements(tpe.(*typex.TType).Type, t.Type().Underlying().(*types.Interface))
	}
	return false
}

func valueOf(v constant.Value) any {
	if v == nil {
		return nil
	}

	switch v.Kind() {
	case constant.Float:
		v, _ := strconv.ParseFloat(v.String(), 10)
		return v
	case constant.Bool:
		v, _ := strconv.ParseBool(v.String())
		return v
	case constant.String:
		v, _ := strconv.Unquote(v.String())
		return v
	case constant.Int:
		v, _ := strconv.ParseInt(v.String(), 10, 64)
		return v
	default:
		return nil
	}
}

func isNil(typ types.Type) bool {
	return typ == nil || typ.String() == types.Typ[types.UntypedNil].String()
}
