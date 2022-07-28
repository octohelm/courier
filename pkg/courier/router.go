package courier

import (
	"bytes"
	"fmt"
	"sort"
	"strings"
)

type Router interface {
	Register(r Router)
	With(routers ...Router) Router
	Routes() Routes
}

func NewRouter(operators ...Operator) Router {
	ops := make([]Operator, 0)
	for i := range operators {
		op := operators[i]

		if withMiddleOperators, ok := op.(WithMiddleOperators); ok {
			ops = append(ops, withMiddleOperators.MiddleOperators()...)
		}

		ops = append(ops, op)
	}

	return &router{
		operators: ops,
	}
}

type router struct {
	parent    *router
	operators []Operator
	children  map[*router]bool
}

func (rt router) With(routers ...Router) Router {
	next := &rt
	for i := range routers {
		next.Register(routers[i])
	}
	return next
}

func (rt *router) Register(r Router) {
	if rt.children == nil {
		rt.children = map[*router]bool{}
	}
	if r.(*router).parent != nil {
		panic(fmt.Errorf("router %v already registered to router %v", r, rt.parent))
	}
	r.(*router).parent = rt
	rt.children[r.(*router)] = true
}

func (rt *router) route() *route {
	parent := rt.parent
	operators := rt.operators

	for parent != nil {
		operators = append(parent.operators, operators...)
		parent = parent.parent
	}

	return &route{
		operators: operators,
		last:      len(rt.children) == 0,
	}
}

func (rt *router) Routes() (routes Routes) {
	maybeAppendRoute := func(router *router) {
		r := router.route()

		if r.last && len(r.operators) > 0 {
			routes = append(routes, r)
		}

		if len(router.children) > 0 {
			routes = append(routes, router.Routes()...)
		}
	}

	if len(rt.children) == 0 {
		maybeAppendRoute(rt)
		return
	}

	for childRouter := range rt.children {
		maybeAppendRoute(childRouter)
	}

	return
}

type Routes []Route

func (routes Routes) String() string {
	keys := make([]string, len(routes))
	for i, r := range routes {
		keys[i] = r.String()
	}
	sort.Strings(keys)
	return strings.Join(keys, "\n")
}

type Route interface {
	fmt.Stringer

	RangeOperator(each func(f *OperatorFactory, i int) error) error
}

type route struct {
	operators []Operator
	last      bool
}

func (r *route) RangeOperator(each func(f *OperatorFactory, i int) error) error {
	lenOfOps := len(r.operators)
	for i, op := range r.operators {
		if err := each(NewOperatorFactory(op, i == lenOfOps-1), i); err != nil {
			return err
		}
	}
	return nil
}

func (r *route) String() string {
	buf := &bytes.Buffer{}
	_ = r.RangeOperator(func(f *OperatorFactory, i int) error {
		if i > 0 {
			buf.WriteString(" |> ")
		}
		buf.WriteString(f.String())
		return nil
	})
	return buf.String()
}
