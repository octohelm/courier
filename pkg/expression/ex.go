package expression

import (
	"context"
	"encoding"
	"fmt"
	"go/ast"
	"reflect"
	"strconv"
	"strings"

	"github.com/pkg/errors"

	"github.com/octohelm/courier/pkg/expression/raw"
)

type Expr interface {
	// Exec final exec
	Exec(ctx context.Context, input any) (any, error)

	String() string
}

type ExprCreator interface {
	// New expr
	New(ctx context.Context, args ...any) (Expr, error)
}

type Expression = []any

func resolveExpression(in any) (string, []any, bool) {
	if expr, ok := in.(Expression); ok {
		if len(expr) >= 1 {
			if name, ok := expr[0].(string); ok {
				return name, expr[1:], ok
			}
		}
	}
	return "", nil, false
}

var defaultFactory = factory{}

type ExprBuilder interface {
	BuildExpression(args ...any) Expression
}

func Register[T any](build func(b ExprBuilder) T, expr Expr) T {
	x := newEx(expr)
	defaultFactory.register(x)
	return build(x)
}

func From(expression Expression) (Expr, error) {
	return defaultFactory.From(expression)
}

type factory map[string]*exprCreator

func (f factory) register(e *exprCreator) {
	f[e.name] = e
}

func (f factory) From(expression Expression) (Expr, error) {
	name, args, ok := resolveExpression(expression)
	if !ok {
		return nil, errors.New("invalid expression, should be [string, ...any]")
	}

	ctx := context.Background()

	return f.from(ctx, name, args)
}

func (f factory) from(ctx context.Context, name string, args []any) (Expr, error) {
	expr, ok := f[name]
	if !ok {
		return nil, errors.Errorf("`%s` is not registered expression", name)
	}

	finalArgs := make([]any, len(args))

	for i := range args {
		arg := args[i]
		n, as, ok := resolveExpression(arg)
		if ok {
			e, err := f.from(ctx, n, as)
			if err != nil {
				return nil, err
			}
			finalArgs[i] = e
			continue
		}
		finalArgs[i] = arg
	}

	return expr.New(ctx, finalArgs...)
}

func Stringify(e any) string {
	switch x := e.(type) {
	case Expression:
		name, args, ok := resolveExpression(x)
		if ok {
			return stringifyExpression(name, args)
		}
		return fmt.Sprintf("%v", x)
	case fmt.Stringer:
		return x.String()
	case string:
		return strconv.Quote(x)
	default:
		return fmt.Sprintf("%v", x)
	}
}

func stringifyExpression(name string, args []any) string {
	b := strings.Builder{}
	b.WriteString(name)
	b.WriteString("(")
	for i := range args {
		if i != 0 {
			b.WriteString(",")
		}
		b.WriteString(Stringify(args[i]))
	}
	b.WriteString(")")
	return b.String()
}

var tpeTextUnmarshalInterface = reflect.TypeOf((*encoding.TextUnmarshaler)(nil)).Elem()
var tpeRawValue = reflect.TypeOf((*raw.Value)(nil)).Elem()

type CanInitFromArgs interface {
	InitFromArgs(args []any) error
}

func newEx(expr Expr) *exprCreator {
	tpe := reflect.TypeOf(expr).Elem()

	e := &exprCreator{
		name: tpe.Name(),
		tpe:  tpe,
	}

	if _, ok := expr.(CanInitFromArgs); ok {
		e.setters = append(e.setters, func(rv reflect.Value, args []any) error {
			return rv.Addr().Interface().(CanInitFromArgs).InitFromArgs(args)
		})
		return e
	}

	argIdx := 0

	for i := 0; i < tpe.NumField(); i++ {
		f := tpe.Field(i)

		if !ast.IsExported(f.Name) {
			return nil
		}

		if f.Type.Name() == "E" {
			if name, ok := f.Tag.Lookup("name"); ok {
				e.name = name
			}
			continue
		}

		if arg, ok := f.Tag.Lookup("arg"); ok {
			if arg == "-" {
				continue
			}

			if arg == "..." {
				e.setters = append(e.setters, func(idx int) func(rv reflect.Value, args []any) error {
					return func(rv reflect.Value, args []any) error {
						rv.Field(idx).Set(reflect.ValueOf(args))
						return nil
					}
				}(i))
				continue
			}
		}

		e.setters = append(e.setters, func(f reflect.StructField, idx int, argIndex int) func(rv reflect.Value, args []any) error {
			if reflect.PointerTo(f.Type).Implements(tpeTextUnmarshalInterface) {
				return func(rv reflect.Value, args []any) error {
					if u, ok := rv.Field(idx).Addr().Interface().(encoding.TextUnmarshaler); ok {
						if err := u.UnmarshalText([]byte(raw.ToString(raw.ValueOf(args[argIndex])))); err != nil {
							return err
						}
					}
					return nil
				}
			}

			if (f.Type).Implements(tpeRawValue) {
				return func(rv reflect.Value, args []any) error {
					if argIndex > len(args) {
						return fmt.Errorf("missing arg %d", argIndex)
					}
					rv.Field(idx).Set(reflect.ValueOf(raw.ValueOf(args[argIndex])))
					return nil
				}
			}

			return func(rv reflect.Value, args []any) error {
				if argIndex > len(args) {
					return fmt.Errorf("missing arg %d", argIndex)
				}
				rv.Field(idx).Set(reflect.ValueOf(args[argIndex]))
				return nil
			}
		}(f, i, argIdx))

		argIdx++
	}

	return e
}

type Setter = func(rv reflect.Value, args []any) error

type exprCreator struct {
	name    string
	tpe     reflect.Type
	setters []Setter
}

func (e *exprCreator) BuildExpression(args ...any) Expression {
	return append([]any{e.name}, args...)
}

func (e *exprCreator) New(ctx context.Context, args ...any) (Expr, error) {
	rv := reflect.New(e.tpe)

	rvv := rv.Elem()
	for i := range e.setters {
		if err := e.setters[i](rvv, args); err != nil {
			return nil, err
		}
	}

	expr := rv.Interface().(Expr)

	if canInit, ok := expr.(interface{ init(name string, args []any) }); ok {
		canInit.init(e.name, args)
	}

	return expr, nil
}

type E struct {
	name string
	args []any
}

func (e *E) init(name string, args []any) {
	e.name = name
	e.args = args
}

func (e *E) String() string {
	return stringifyExpression(e.name, e.args)
}
