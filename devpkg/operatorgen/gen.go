package operatorgen

import (
	"fmt"
	"go/ast"
	"go/constant"
	"go/types"
	"reflect"
	"strconv"
	"strings"

	"github.com/octohelm/courier/pkg/statuserror"

	"github.com/octohelm/courier/pkg/courier"
	"github.com/octohelm/courier/pkg/courierhttp"
	"github.com/octohelm/gengo/pkg/gengo"
	gengotypes "github.com/octohelm/gengo/pkg/types"
	typex "github.com/octohelm/x/types"
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
					c.Logger().Warn(fmt.Errorf("%s return multi types, `%s` `%s`", named, tpe, tpe2))
				}
			}

			tpe = tpe2
			expr = resp.Expr

			// use first got type
			break
		}
	}

	if isNil(tpe) {
		c.Render(gengo.Snippet{
			gengo.T: `
func (*@Type) ResponseContent() any {
	return nil
}

`,
			"Type": gengo.ID(named.Obj()),
		})

	} else if types.IsInterface(tpe) && !strings.Contains(tpe.String(), "github.com/octohelm/courier/pkg/courierhttp.Response") {
		c.Logger().Warn(fmt.Errorf("%s return interface %s will be untyped jsonschema", named, tpe))
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
	return new(@ReturnType)
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
	tags, _ := c.Doc(named.Obj())

	if registers, ok := tags["gengo:operator:register"]; ok {
		for _, register := range registers {
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
	}
}

func (g *operatorGen) resolvePkg(c gengo.Context, importPath string) *types.Package {
	return c.Package(importPath).Pkg()
}

func (g *operatorGen) firstValueOfFunc(c gengo.Context, named *types.Named, name string) (interface{}, bool) {
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
	default:
		return nil
	}
}

func isNil(typ types.Type) bool {
	return typ == nil || typ.String() == types.Typ[types.UntypedNil].String()
}
