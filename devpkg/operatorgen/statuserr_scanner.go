package openapi

import (
	"fmt"
	"go/ast"
	"go/types"
	"reflect"
	"sort"
	"strconv"

	"github.com/octohelm/courier/pkg/statuserror"
	"github.com/octohelm/gengo/pkg/gengo"
	gengotypes "github.com/octohelm/gengo/pkg/types"
	"github.com/pkg/errors"
)

func newStatusErrScanner() *statusErrScanner {
	return &statusErrScanner{
		statusErrorTypes: map[*types.Named][]*statuserror.StatusErr{},
		errorsUsed:       map[*types.Func][]*statuserror.StatusErr{},
	}
}

type statusErrScanner struct {
	statusErrorTypes map[*types.Named][]*statuserror.StatusErr
	errorsUsed       map[*types.Func][]*statuserror.StatusErr
}

var statusErr = reflect.TypeOf(statuserror.StatusErr{})

func isTypeStatusErr(named *types.Named) bool {
	if obj := named.Obj(); obj != nil {
		if pkg := obj.Pkg(); pkg != nil {
			return pkg.Path() == statusErr.PkgPath() && obj.Name() == statusErr.Name()
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

func (s *statusErrScanner) StatusErrorsInFunc(ctx gengo.Context, typeFunc *types.Func) []*statuserror.StatusErr {
	if typeFunc == nil {
		return nil
	}

	if statusErrList, ok := s.errorsUsed[typeFunc]; ok {
		return statusErrList
	}

	s.errorsUsed[typeFunc] = []*statuserror.StatusErr{}

	pkg := ctx.Package(typeFunc.Pkg().Path())

	_, lines := pkg.Doc(typeFunc.Pos())
	s.appendStateErrs(typeFunc, pickStatusErrorsFromDoc(lines)...)

	results, n := pkg.ResultsOf(typeFunc)
	for i := 0; i < n; i++ {
		for _, r := range results[i] {
			tpe := r.Type
			if p, ok := tpe.(*types.Pointer); ok {
				tpe = p.Elem()
			}
			if named, ok := tpe.(*types.Named); ok {
				if isTypeStatusErr(named) {
					ast.Inspect(r.Expr, func(node ast.Node) bool {
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

	return s.errorsUsed[typeFunc]
}

func (s *statusErrScanner) appendStateErrs(typeFunc *types.Func, statusErrs ...*statuserror.StatusErr) {
	m := map[string]*statuserror.StatusErr{}

	errs := append(s.errorsUsed[typeFunc], statusErrs...)
	for i := range errs {
		s := errs[i]
		m[fmt.Sprintf("%s%d", s.Key, s.Code)] = s
	}

	next := make([]*statuserror.StatusErr, 0)
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

		code := 0
		key := ""
		msg := ""
		desc := make([]string, 0)

		for i, arg := range x.Args[1:] {
			tv, err := pkg.Eval(arg)
			if err != nil {
				continue
			}

			switch i {
			case 0: // code
				code, _ = strconv.Atoi(tv.Value.String())
			case 1: // key
				key, _ = strconv.Unquote(tv.Value.String())
			case 2: // msg
				msg, _ = strconv.Unquote(tv.Value.String())
			default:
				d, _ := strconv.Unquote(tv.Value.String())
				desc = append(desc, d)
			}
		}

		if code > 0 {
			if msg == "" {
				msg = key
			}

			s.appendStateErrs(typeFunc, statuserror.Wrap(errors.New(""), code, key, append([]string{msg}, desc...)...))
		}

		return true
	}

	return false
}

func pickStatusErrorsFromDoc(lines []string) []*statuserror.StatusErr {
	statusErrorList := make([]*statuserror.StatusErr, 0)

	for _, line := range lines {
		if line != "" {
			if statusErr, err := statuserror.ParseStatusErrSummary(line); err == nil {
				statusErrorList = append(statusErrorList, statusErr)
			}
		}
	}

	return statusErrorList
}
