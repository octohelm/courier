package expression

import (
	"context"
	"fmt"
	"testing"

	. "github.com/octohelm/x/testing/v2"

	"github.com/octohelm/courier/pkg/expression/raw"
)

func TestCompositeExpressions(t *testing.T) {
	t.Run("array every", func(t *testing.T) {
		e, _ := From(Every(
			Elem(Pipe(Len(), Gte(3))),
		))

		t.Run("returns true when every element matches", func(t *testing.T) {
			ret, _ := e.Exec(context.Background(), []any{"123", "123", "123"})
			expectExpr(t, ret, true)
		})

		t.Run("returns false when some element is too short", func(t *testing.T) {
			ret, _ := e.Exec(context.Background(), []any{"123", "11"})
			expectExpr(t, ret, false)
		})

		t.Run("returns false when any element fails", func(t *testing.T) {
			ret, _ := e.Exec(context.Background(), []any{"123", "11", "123"})
			expectExpr(t, ret, false)
		})
	})

	t.Run("array some", func(t *testing.T) {
		e, _ := From(Some(
			Elem(Pipe(Len(), Gte(3))),
		))

		t.Run("returns true when all elements match", func(t *testing.T) {
			ret, _ := e.Exec(context.Background(), []any{"123", "123", "123"})
			expectExpr(t, ret, true)
		})

		t.Run("returns true when some element matches", func(t *testing.T) {
			ret, _ := e.Exec(context.Background(), []any{"123", "11", "1"})
			expectExpr(t, ret, true)
		})

		t.Run("returns false when no element matches", func(t *testing.T) {
			ret, _ := e.Exec(context.Background(), []any{"12", "11", "1"})
			expectExpr(t, ret, false)
		})
	})

	t.Run("eq", func(t *testing.T) {
		ex, err := From(Eq(1))
		expectErrNil(t, err)
		expectExpr(t, ex.String(), "eq(1)")

		ret, err := ex.Exec(context.Background(), 1)
		expectErrNil(t, err)
		expectExpr(t, ret, true)
	})

	t.Run("match", func(t *testing.T) {
		e, err := From(Match("[a-z]+"))
		expectErrNil(t, err)
		expectExpr(t, e.String(), `match("[a-z]+")`)

		ret, err := e.Exec(context.Background(), "abc")
		expectErrNil(t, err)
		expectExpr(t, ret, true)
	})

	t.Run("pipe", func(t *testing.T) {
		e, err := From(Pipe(Get("x"), Len(), Eq(5)))
		expectErrNil(t, err)

		ctx := WithValueGetter(context.Background(), ValueGetterFunc(func(name string) (any, bool) {
			return "12345", true
		}))

		ret, err := e.Exec(ctx, nil)
		expectErrNil(t, err)
		expectExpr(t, ret, true)
	})

	t.Run("simple when", func(t *testing.T) {
		ex, _ := From(
			When(
				Pipe(Get("x"), Eq(1)),
				Pipe("x"),
			),
		)

		t.Run("returns branch value when condition matches", func(t *testing.T) {
			ctx := WithValueGetter(context.Background(), ValueGetterFunc(func(name string) (any, bool) {
				return 1, true
			}))
			ret, err := ex.Exec(ctx, nil)
			expectErrNil(t, err)
			expectExpr(t, ret, "x")
		})

		t.Run("returns nil when condition does not match", func(t *testing.T) {
			ctx := WithValueGetter(context.Background(), ValueGetterFunc(func(name string) (any, bool) {
				return 2, true
			}))
			ret, err := ex.Exec(ctx, nil)
			expectErrNil(t, err)
			expectExpr[any](t, ret, nil)
		})
	})

	t.Run("simple select", func(t *testing.T) {
		ex, _ := From(
			Select(
				When(
					Pipe(Get("x"), Eq(1)),
					Eq(1),
				),
				Eq(2),
			),
		)

		t.Run("returns first branch when condition matches", func(t *testing.T) {
			ctx := WithValueGetter(context.Background(), ValueGetterFunc(func(name string) (any, bool) {
				return 1, true
			}))
			ret, err := ex.Exec(ctx, 1)
			expectErrNil(t, err)
			expectExpr(t, ret, true)
		})

		t.Run("falls back to default branch when condition does not match", func(t *testing.T) {
			ctx := WithValueGetter(context.Background(), ValueGetterFunc(func(name string) (any, bool) {
				return 2, true
			}))

			{
				ret, err := ex.Exec(ctx, 1)
				expectErrNil(t, err)
				expectExpr(t, ret, false)
			}

			{
				ret, err := ex.Exec(ctx, 2)
				expectErrNil(t, err)
				expectExpr(t, ret, true)
			}
		})
	})
}

type fixture struct {
	in  any
	ret any
}

var cases = []struct {
	summary  string
	ex       Expression
	fixtures []fixture
}{
	{
		summary: "str len",
		ex:      Pipe(Len(), Eq(5)),
		fixtures: []fixture{
			{"12345", true},
			{"123", false},
		},
	},
	{
		summary: "charCount",
		ex:      Pipe(CharCount(), Eq(1)),
		fixtures: []fixture{
			{"😀", true},
			{"😀😀", false},
		},
	},
	{
		summary: "eq",
		ex:      Eq(1.0),
		fixtures: []fixture{
			{1, true},
			{2, false},
		},
	},
	{
		summary: "lt",
		ex:      Lt(10),
		fixtures: []fixture{
			{1, true},
			{10, false},
			{11, false},
		},
	},
	{
		summary: "lte",
		ex:      Lte(10),
		fixtures: []fixture{
			{1, true},
			{10, true},
			{11, false},
		},
	},
	{
		summary: "gt",
		ex:      Gt(10),
		fixtures: []fixture{
			{11, true},
			{10, false},
			{1, false},
		},
	},
	{
		summary: "gte",
		ex:      Gte(10),
		fixtures: []fixture{
			{11, true},
			{10, true},
			{1, false},
		},
	},
	{
		summary: "match",
		ex:      Match("[a-z]+"),
		fixtures: []fixture{
			{"abc", true},
			{"ABC", false},
		},
	},
	{
		summary: "allOf",
		ex:      AllOf(Gt(3), Lt(5)),
		fixtures: []fixture{
			{3, false},
			{5, false},
			{4, true},
		},
	},
	{
		summary: "oneOf as or",
		ex: OneOf(
			Lte(3),
			Gte(5),
		),
		fixtures: []fixture{
			{4, false},
			{5, true},
			{3, true},
		},
	},
	{
		summary: "not",
		ex:      Not(Lte(10)),
		fixtures: []fixture{
			{11, true},
			{10, false},
			{1, false},
		},
	},
	{
		summary: "oneOf as enum",
		ex:      OneOf(1, 3, 5),
		fixtures: []fixture{
			{4, false},
			{3, true},
		},
	},
}

func TestExpressionFixtures(t *testing.T) {
	for i := range cases {
		c := cases[i]

		ex, err := From(c.ex)
		expectErrNil(t, err)

		t.Run(fmt.Sprintf("%s: %s", c.summary, ex.String()), func(t *testing.T) {
			for j := range c.fixtures {
				ft := c.fixtures[j]

				t.Run(fmt.Sprintf("x(%v)=%v", ft.in, ft.ret), func(t *testing.T) {
					out, err := ex.Exec(context.Background(), ft.in)
					expectErrNil(t, err)
					expectExpr(t, out, ft.ret)
				})
			}
		})
	}
}

func Benchmark(b *testing.B) {
	ex, _ := From(Eq(1))

	b.Run("expr", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = ex.Exec(context.Background(), 1)
		}
	})

	b.Run("direct", func(b *testing.B) {
		exec := func(ctx context.Context, v any) (any, error) {
			return raw.Compare(raw.ValueOf(v), raw.ValueOf(1))
		}

		for i := 0; i < b.N; i++ {
			_, _ = exec(context.Background(), 1)
		}
	})
}

func expectExpr[V any](t *testing.T, actual V, expected V) {
	Then(t, "表达式结果符合预期", Expect(actual, Equal(expected)))
}

func expectErrNil(t *testing.T, err error) {
	Then(t, "执行过程中不返回错误", Expect(err, Equal[error](nil)))
}
