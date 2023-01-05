package expression

import (
	"context"
	"fmt"
	"testing"

	"github.com/octohelm/courier/pkg/expression/raw"

	testingx "github.com/octohelm/x/testing"
)

func TestExpr(t *testing.T) {
	t.Run("array", func(t *testing.T) {
		e, _ := From(AllOf(
			Pipe(Len(), Gte(3)),
			Each(
				Elem(Pipe(Len(), Gte(3))),
			),
		))

		t.Run("should pass", func(t *testing.T) {
			ret, _ := e.Exec(context.Background(), []any{"123", "123", "123"})
			testingx.Expect(t, ret, testingx.Be[any](true))
		})

		t.Run("should failed cause array len", func(t *testing.T) {
			ret, _ := e.Exec(context.Background(), []any{"123", "11"})
			testingx.Expect(t, ret, testingx.Be[any](false))
		})

		t.Run("should failed cause elem len", func(t *testing.T) {
			ret, _ := e.Exec(context.Background(), []any{"123", "11", "123"})
			testingx.Expect(t, ret, testingx.Be[any](false))
		})
	})

	t.Run("eq", func(t *testing.T) {
		ex, err := From(Eq(1))
		testingx.Expect(t, err, testingx.Be[error](nil))
		testingx.Expect(t, ex.String(), testingx.Be[string]("eq(1)"))

		ret, err := ex.Exec(context.Background(), 1)
		testingx.Expect(t, err, testingx.Be[error](nil))
		testingx.Expect(t, ret, testingx.Be[any](true))
	})

	t.Run("match", func(t *testing.T) {
		e, err := From(Match("[a-z]+"))
		testingx.Expect(t, err, testingx.Be[error](nil))
		testingx.Expect(t, e.String(), testingx.Be[string](`match("[a-z]+")`))

		ret, err := e.Exec(context.Background(), "abc")
		testingx.Expect(t, err, testingx.Be[error](nil))
		testingx.Expect(t, ret, testingx.Be[any](true))
	})

	t.Run("pipe", func(t *testing.T) {
		e, err := From(Pipe(Get("x"), Len(), Eq(5)))

		testingx.Expect(t, err, testingx.Be[error](nil))

		ctx := WithValueGetter(context.Background(), ValueGetterFunc(func(name string) (any, bool) {
			return "12345", true
		}))

		ret, err := e.Exec(ctx, nil)
		testingx.Expect(t, err, testingx.Be[error](nil))
		testingx.Expect(t, ret, testingx.Be[any](true))
	})

	t.Run("simple when", func(t *testing.T) {
		ex, _ := From(
			When(
				Pipe(Get("x"), Eq(1)),
				Pipe("x"),
			),
		)

		t.Run("when condition pass", func(t *testing.T) {
			ctx := WithValueGetter(context.Background(), ValueGetterFunc(func(name string) (any, bool) {
				return 1, true
			}))
			ret, err := ex.Exec(ctx, nil)
			testingx.Expect(t, err, testingx.Be[error](nil))
			testingx.Expect(t, ret, testingx.Be[any]("x"))
		})

		t.Run("when condition not pass", func(t *testing.T) {
			ctx := WithValueGetter(context.Background(), ValueGetterFunc(func(name string) (any, bool) {
				return 2, true
			}))
			ret, err := ex.Exec(ctx, nil)
			testingx.Expect(t, err, testingx.Be[error](nil))
			testingx.Expect(t, ret, testingx.Be[any](nil))
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

		t.Run("when condition pass", func(t *testing.T) {
			ctx := WithValueGetter(context.Background(), ValueGetterFunc(func(name string) (any, bool) {
				return 1, true
			}))
			ret, err := ex.Exec(ctx, 1)
			testingx.Expect(t, err, testingx.Be[error](nil))
			testingx.Expect(t, ret, testingx.Be[any](true))
		})

		t.Run("when condition not pass", func(t *testing.T) {
			ctx := WithValueGetter(context.Background(), ValueGetterFunc(func(name string) (any, bool) {
				return 2, true
			}))

			{
				ret, err := ex.Exec(ctx, 1)
				testingx.Expect(t, err, testingx.Be[error](nil))
				testingx.Expect(t, ret, testingx.Be[any](false))
			}

			{
				ret, err := ex.Exec(ctx, 2)
				testingx.Expect(t, err, testingx.Be[error](nil))
				testingx.Expect(t, ret, testingx.Be[any](true))
			}
		})
	})
}

type fixture struct {
	in  any
	ret any
}

var (
	cases = []struct {
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
)

func TestExpressions(t *testing.T) {
	for i := range cases {
		c := cases[i]

		ex, err := From(c.ex)
		testingx.Expect(t, err, testingx.Be[error](nil))

		t.Run(fmt.Sprintf("%s: %s", c.summary, ex.String()), func(t *testing.T) {
			for j := range c.fixtures {
				ft := c.fixtures[j]

				t.Run(fmt.Sprintf("x(%v)=%v", ft.in, ft.ret), func(t *testing.T) {
					out, err := ex.Exec(context.Background(), ft.in)
					testingx.Expect(t, err, testingx.Be[error](nil))
					testingx.Expect(t, out, testingx.Be[any](ft.ret))
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
