package openapi

import (
	"fmt"
	"go/ast"
	"go/constant"
	"go/types"
	"reflect"
	"strconv"

	"github.com/octohelm/courier/pkg/statuserror"

	"github.com/octohelm/courier/pkg/courier"
	"github.com/octohelm/courier/pkg/courierhttp"
	"github.com/octohelm/gengo/pkg/gengo"
	gengotypes "github.com/octohelm/gengo/pkg/types"
	typesutil "github.com/octohelm/x/types"
)

func init() {
	gengo.Register(&operatorGen{})
}

type operatorGen struct {
}

func (g *operatorGen) Name() string {
	return "operator"
}

var statusErrorScanner = newStatusErrScanner()

func (g *operatorGen) GenerateType(c gengo.Context, named *types.Named) error {
	if !ast.IsExported(named.Obj().Name()) {
		return gengo.ErrSkip
	}

	if !isCourierOperator(c, typesutil.FromTType(types.NewPointer(named)), g.resolvePkg) {
		return gengo.ErrSkip
	}

	g.generateRegister(c, named)
	g.generateReturns(c, named)
	return nil
}

func (g *operatorGen) generateReturns(c gengo.Context, named *types.Named) {
	method, ok := typesutil.FromTType(types.NewPointer(named)).MethodByName("Output")
	if ok {
		results, n := c.Package(named.Obj().Pkg().Path()).ResultsOf(method.(*typesutil.TMethod).Func)
		if n == 2 {
			g.generateSuccessReturn(c, named, results[0])
			g.generateErrorsReturn(c, named, method.(*typesutil.TMethod).Func)
		}
	}
}

func (g *operatorGen) generateErrorsReturn(c gengo.Context, named *types.Named, fn *types.Func) {
	statusErrors := statusErrorScanner.StatusErrorsInFunc(c, fn)
	if len(statusErrors) > 0 {
		c.Render(gengo.Snippet{
			gengo.T: `
func (*@Type) ResponseErrors() []error {
	return []error{
		@statusErrors
	}
}

`,
			"Type": gengo.ID(named.Obj()),
			"statusErrors": gengo.MapSnippet(statusErrors, func(statusError *statuserror.StatusErr) gengo.Snippet {
				return gengo.Snippet{
					gengo.T:       "@statusError,",
					"statusError": statusError,
				}
			}),
		})
	}
}

func (g *operatorGen) generateSuccessReturn(c gengo.Context, named *types.Named, typeAndValues gengotypes.TypeAndValues) {
	var tpe types.Type
	var expr ast.Expr

	for _, resp := range typeAndValues {
		if resp.Type != nil {
			tpe2 := dePtr(resp.Type)

			if isNil(tpe2) {
				continue
			}

			if !isNil(tpe) {
				if tpe.String() != tpe2.String() {
					panic(fmt.Errorf("%s return multi types, `%s` `%s`", named, tpe, tpe2))
				}
			}

			tpe = tpe2
			expr = resp.Expr
		}
	}

	if isNil(tpe) || isAny(tpe) {
		c.Render(gengo.Snippet{
			gengo.T: `
func (*@Type) ResponseContent() any {
	return nil
}

`,
			"Type": gengo.ID(named.Obj()),
		})

	} else {
		if n, ok := tpe.(*types.Named); ok {
			typeArgs := n.TypeArgs()

			if typeArgs.Len() > 0 {
				if n.Obj().Pkg().Path() == typeResponseWithSettingPkgPath.PkgPath() && n.Obj().Name() == "Response" {
					tpe = dePtr(n.TypeArgs().At(0))

					ast.Inspect(expr, func(node ast.Node) bool {
						switch callExpr := node.(type) {
						case *ast.CallExpr:
							switch e := callExpr.Fun.(type) {
							case *ast.SelectorExpr:
								switch e.Sel.Name {
								case "WithStatusCode", "Redirect":
									if p := c.LocateInPackage(node.Pos()); p != nil {
										v, _ := p.Eval(callExpr.Args[0])
										if statueCode, ok := valueOf(v.Value).(int64); ok {
											c.Render(gengo.Snippet{gengo.T: `
func (*@Type) ResponseStatusCode() int {
	return @statueCode
}

`,
												"Type":       gengo.ID(named.Obj()),
												"statueCode": int(statueCode),
											})
										}
									}
									return false
								case "WithContentType":
									if p := c.LocateInPackage(node.Pos()); p != nil {
										v, _ := p.Eval(callExpr.Args[0])
										if contentType, ok := valueOf(v.Value).(string); ok {
											c.Render(gengo.Snippet{gengo.T: `
func (*@Type) ResponseContentType() string {
	return @contentType
}

`,
												"Type":        gengo.ID(named.Obj()),
												"contentType": contentType,
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

		if _, ok := tpe.(*types.Interface); ok {
			return
		}

		c.Render(gengo.Snippet{gengo.T: `
func (*@Type) ResponseContent() any {
	return &@ReturnType{}
}

`,
			"Type":       gengo.ID(named.Obj()),
			"ReturnType": gengo.ID(tpe),
		})
	}
}

func dePtr(t types.Type) types.Type {
	if p, ok := t.(*types.Pointer); ok {
		t = p.Elem()
	}
	return t
}

var typeResponseWithSettingPkgPath = reflect.TypeOf((*courierhttp.Response[any])(nil)).Elem()

func (g *operatorGen) generateRegister(c gengo.Context, named *types.Named) {
	register := ""

	tags, _ := c.Doc(named.Obj())

	if r, ok := tags["gengo:operator:register"]; ok {
		if len(r) > 0 {
			register = r[0]
		}
	}

	if register == "" {
		return
	}

	c.Render(gengo.Snippet{gengo.T: `
func init() {
	@R.Register(@courierNewRouter(&@Operator{}))
}

`,
		"R":                gengo.ID(register),
		"courierNewRouter": gengo.ID("github.com/octohelm/courier/pkg/courier.NewRouter"),
		"Operator":         gengo.ID(named.Obj()),
	})
}

func (g *operatorGen) resolvePkg(c gengo.Context, importPath string) *types.Package {
	return c.Package(importPath).Pkg()
}

func (g *operatorGen) firstValueOfFunc(c gengo.Context, named *types.Named, name string) (interface{}, bool) {
	method, ok := typesutil.FromTType(types.NewPointer(named)).MethodByName(name)
	if ok {
		fn := method.(*typesutil.TMethod).Func
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

func isCourierOperator(c gengo.Context, tpe typesutil.Type, lookup func(c gengo.Context, importPath string) *types.Package) bool {
	switch tpe.(type) {
	case *typesutil.RType:
		return tpe.Implements(typesutil.FromRType(typOperator))
	case *typesutil.TType:
		pkg := lookup(c, typOperator.PkgPath())
		if pkg == nil {
			return false
		}
		t := pkg.Scope().Lookup(typOperator.Name())
		if t == nil {
			return false
		}
		return types.Implements(tpe.(*typesutil.TType).Type, t.Type().Underlying().(*types.Interface))
	}
	return false
}

func valueOf(v constant.Value) interface{} {
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
	}

	return nil
}

func isNil(typ types.Type) bool {
	return typ == nil || typ.String() == types.Typ[types.UntypedNil].String()
}

func isAny(typ types.Type) bool {
	return types.IsInterface(typ) && typ.String() == "interface {}"
}
