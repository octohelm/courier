package operatorgen

import (
	"fmt"
	"go/ast"
	"go/constant"
	"go/token"
	"go/types"
	"net/http"
	"reflect"
	"sort"
	"strconv"

	"github.com/octohelm/courier/pkg/statuserror"
	"github.com/octohelm/gengo/pkg/gengo"
	gengotypes "github.com/octohelm/gengo/pkg/types"
	typex "github.com/octohelm/x/types"
)

func newStatusErrScanner() *statusErrScanner {
	return &statusErrScanner{
		statusErrorTypes: map[*types.Named][]*statuserror.Descriptor{},
		errorsUsed:       map[*types.Func][]*statuserror.Descriptor{},
	}
}

type statusErrScanner struct {
	statusErrorTypes map[*types.Named][]*statuserror.Descriptor
	errorsUsed       map[*types.Func][]*statuserror.Descriptor
}

var statusErr = reflect.TypeOf(statuserror.Descriptor{})

func isTypeStatusErr(named *types.Named) bool {
	if o := named.Obj(); o != nil {
		if pkg := o.Pkg(); pkg != nil {
			return pkg.Path() == statusErr.PkgPath() && o.Name() == statusErr.Name()
		}
	}
	return false
}

func identChainOfCallFunc(expr ast.Expr) (list []*ast.Ident) {
	switch e := expr.(type) {
	case *ast.CallExpr:
		list = append(list, identChainOfCallFunc(e.Fun)...)
	case *ast.SelectorExpr:
		list = append(list, identChainOfCallFunc(e.X)...)
		list = append(list, e.Sel)
	case *ast.Ident:
		list = append(list, e)
	}
	return
}

func (s *statusErrScanner) StatusErrorsInFunc(ctx gengo.Context, typeFunc *types.Func) []*statuserror.Descriptor {
	if typeFunc == nil {
		return nil
	}

	if statusErrList, ok := s.errorsUsed[typeFunc]; ok {
		return statusErrList
	}

	s.errorsUsed[typeFunc] = []*statuserror.Descriptor{}

	pkg := ctx.Package(typeFunc.Pkg().Path())

	results, n := pkg.ResultsOf(typeFunc)
	if n > 0 {
		for _, ret := range results[n-1] {
			if named, ok := dePtr(ret.Type).(*types.Named); ok {
				if !ast.IsExported(named.Obj().Name()) {
					continue
				}

				if isErrWithStatusCodeInterface(named) {
					s.appendStateErrs(typeFunc, s.scanErrWithStatusCodeInterface(ctx, named)...)
					continue
				}

				if isTypeStatusErr(named) {
					if ret.Expr != nil {
						ast.Inspect(ret.Expr, func(node ast.Node) bool {
							switch x := node.(type) {
							case *ast.CallExpr:
								identList := identChainOfCallFunc(x.Fun)

								if len(identList) > 0 {
									callIdent := identList[len(identList)-1]
									obj := pkg.ObjectOf(callIdent)

									if obj != nil {
										if ok := s.scanStatusErrIsExist(typeFunc, pkg, obj, callIdent, x); ok {
											return true
										}

										if nextFuncType, ok := obj.(*types.Func); ok && nextFuncType != typeFunc && nextFuncType.Pkg() != nil {
											s.appendStateErrs(typeFunc, s.StatusErrorsInFunc(ctx, nextFuncType)...)
										}
									}
								}
							}
							return true
						})
					}
				}
			}
		}
	}

	return s.errorsUsed[typeFunc]
}

func (s *statusErrScanner) appendStateErrs(typeFunc *types.Func, statusErrs ...*statuserror.Descriptor) {
	m := map[string]*statuserror.Descriptor{}

	errs := append(s.errorsUsed[typeFunc], statusErrs...)
	for i := range errs {
		s := errs[i]
		m[fmt.Sprintf("%s%d", s.Code, s.Status)] = s
	}

	next := make([]*statuserror.Descriptor, 0)
	for k := range m {
		next = append(next, m[k])
	}

	sort.Slice(next, func(i, j int) bool {
		return next[i].Code < next[j].Code
	})

	s.errorsUsed[typeFunc] = next
}

func (s *statusErrScanner) scanStatusErrIsExist(typeFunc *types.Func, pkg gengotypes.Package, obj types.Object, callIdent *ast.Ident, x *ast.CallExpr) bool {
	if callIdent.Name == "Wrap" && obj.Pkg().Path() == statusErr.PkgPath() {
		statusCode := 0
		code := ""
		msg := ""
		desc := make([]string, 0)

		for i, arg := range x.Args[1:] {
			tv, err := pkg.Eval(arg)
			if err != nil {
				continue
			}

			if tv.Value == nil {
				continue
			}

			switch i {
			case 0: // statusCode
				statusCode, _ = strconv.Atoi(tv.Value.String())
			case 1: // code
				code, _ = strconv.Unquote(tv.Value.String())
			case 2: // msg
				msg, _ = strconv.Unquote(tv.Value.String())
			default:
				d, _ := strconv.Unquote(tv.Value.String())
				desc = append(desc, d)
			}
		}

		if statusCode > 0 {
			if msg == "" {
				msg = code
			}

			s.appendStateErrs(typeFunc, &statuserror.Descriptor{
				Code:    code,
				Status:  statusCode,
				Message: msg,
			})
		}

		return true
	}

	return false
}

var rtypeErrorWithStatusCode = typex.FromRType(reflect.TypeOf((*statuserror.WithStatusCode)(nil)).Elem())

func isErrWithStatusCodeInterface(named *types.Named) bool {
	if named != nil {
		return typex.FromTType(types.NewPointer(named)).Implements(rtypeErrorWithStatusCode)
	}
	return false
}

func (s *statusErrScanner) resolveStateCode(ctx gengo.Context, named *types.Named) (int, bool) {
	method, ok := typex.FromTType(types.NewPointer(named)).MethodByName("StatusCode")
	if ok {
		m := method.(*typex.TMethod)
		if m.Func.Pkg() == nil {
			return 0, false
		}

		results, n := ctx.Package(m.Func.Pkg().Path()).ResultsOf(m.Func)
		if n == 1 {
			for _, r := range results[0] {
				if r.Value != nil && r.Value.Kind() == constant.Int {
					v, err := strconv.ParseInt(r.Value.String(), 10, 64)
					if err == nil {
						return int(v), true
					}
				}
			}
		}
	}
	return 0, false
}

func (s *statusErrScanner) scanErrWithStatusCodeInterface(ctx gengo.Context, named *types.Named) (list []*statuserror.Descriptor) {
	if named.Obj() == nil {
		return nil
	}

	serr := &statuserror.Descriptor{
		Code:   named.Obj().Pkg().Path() + "." + named.Obj().Name(),
		Status: http.StatusInternalServerError,
	}

	code, ok := s.resolveStateCode(ctx, named)
	if ok {
		serr.Status = code
	}

	method, ok := typex.FromTType(types.NewPointer(named)).MethodByName("Error")
	if ok {
		m := method.(*typex.TMethod)
		if m.Func.Pkg() == nil {
			return
		}

		results, n := ctx.Package(m.Func.Pkg().Path()).ResultsOf(m.Func)
		if n == 1 {
			for _, r := range results[0] {
				switch x := r.Expr.(type) {
				case *ast.BasicLit:
					str, err := strconv.Unquote(x.Value)
					if err == nil {
						e := &(*serr)
						e.Message = str
						list = append(list, e)
					}
				case *ast.CallExpr:
					if selectExpr, ok := x.Fun.(*ast.SelectorExpr); ok {
						if selectExpr.Sel.Name == "Sprintf" {
							e := &(*serr)
							e.Message = fmtSprintfArgsAsTemplate(x.Args)
							list = append(list, e)
						}
					}
				}
			}
		}
	}

	return
}

func fmtSprintfArgsAsTemplate(args []ast.Expr) string {
	if len(args) == 0 {
		return ""
	}

	f := ""
	fArgs := make([]any, 0, len(args))

	toString := func(a *ast.BasicLit) string {
		switch a.Kind {
		case token.STRING:
			v, _ := strconv.Unquote(a.Value)
			return v
		default:
			return a.Value
		}
	}

	for i, arg := range args {
		switch a := arg.(type) {
		case *ast.BasicLit:
			if i == 0 {
				f = toString(a)
			} else {
				fArgs = append(fArgs, toString(a))
			}
		case *ast.SelectorExpr:
			fArgs = append(fArgs, fmt.Sprintf("{%s}", a.Sel.Name))
		case *ast.Ident:
			fArgs = append(fArgs, fmt.Sprintf("{%s}", a.Name))
		}
	}

	return fmt.Sprintf(normalizeFormat(f), fArgs...)
}
