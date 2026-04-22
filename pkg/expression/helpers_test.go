package expression

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	. "github.com/octohelm/x/testing/v2"
)

func TestExpressionHelpersAndErrors(t0 *testing.T) {
	Then(t0, "公开辅助方法与错误分支符合预期",
		ExpectMust(func() error {
			name, args, ok := resolveExpression([]any{"eq", 1})
			if name != "eq" || len(args) != 1 || args[0] != 1 || !ok {
				return fmt.Errorf("unexpected resolve result %q %#v %v", name, args, ok)
			}
			return nil
		}),
		Expect(Stringify(Pipe("x", Eq(1))), Equal(`pipe("x",eq(1))`)),
		Expect(Stringify("demo"), Equal(`"demo"`)),
		Expect(Stringify(1), Equal("1")),
		ExpectMust(func() error {
			expr, err := From([]any{1})
			if err == nil || expr != nil {
				return fmt.Errorf("expected invalid expression error, got %v %v", expr, err)
			}
			return nil
		}),
		ExpectMust(func() error {
			expr, err := From([]any{"missing"})
			if err == nil || expr != nil {
				return fmt.Errorf("expected missing expression error, got %v %v", expr, err)
			}
			return nil
		}),
		ExpectMust(func() error {
			expr, err := From(Match("["))
			if err == nil || expr != nil {
				return fmt.Errorf("expected invalid regexp error, got %v %v", expr, err)
			}
			return nil
		}),
	)
}

func TestComposeAndContextBranches(t0 *testing.T) {
	Then(t0, "组合表达式与上下文回退链可正常工作",
		ExpectMust(func() error {
			expr, err := From(Some(Key(Eq("name"))))
			if err != nil {
				return err
			}
			out, err := expr.Exec(context.Background(), map[string]any{"name": "demo"})
			if err != nil {
				return err
			}
			if out != true {
				return fmt.Errorf("unexpected some result %v", out)
			}
			return nil
		}),
		ExpectMust(func() error {
			expr, err := From(Every(Key(Match("^x-"))))
			if err != nil {
				return err
			}
			out, err := expr.Exec(context.Background(), map[string]any{"x-a": 1, "x-b": 2})
			if err != nil {
				return err
			}
			if out != true {
				return fmt.Errorf("unexpected every result %v", out)
			}
			return nil
		}),
		ExpectDo(func() error {
			expr, _ := From(Elem(Eq(1)))
			_, err := expr.Exec(context.Background(), 1)
			return err
		}, ErrorMatch(mustRE("`elem` must be used"))),
		ExpectDo(func() error {
			expr, _ := From(Key(Eq("a")))
			_, err := expr.Exec(context.Background(), "a")
			return err
		}, ErrorMatch(mustRE("`key` must be used"))),
		ExpectDo(func() error {
			expr, _ := From(Get("x"))
			_, err := expr.Exec(context.Background(), nil)
			return err
		}, ErrorMatch(mustRE("missing value getter"))),
		ExpectDo(func() error {
			expr, _ := From(Get("missing"))
			ctx := WithValueGetter(context.Background(), ValueGetterFunc(func(name string) (any, bool) {
				return nil, false
			}))
			_, err := expr.Exec(ctx, nil)
			return err
		}, ErrorMatch(mustRE("missing field missing"))),
		ExpectMust(func() error {
			ctx := WithValueGetter(
				WithValueGetter(context.Background(), ValueGetterFunc(func(name string) (any, bool) {
					if name == "base" {
						return "base-value", true
					}
					return nil, false
				})),
				ValueGetterFunc(func(name string) (any, bool) {
					if name == "child" {
						return "child-value", true
					}
					return nil, false
				}),
			)

			childExpr, _ := From(Get("child"))
			baseExpr, _ := From(Get("base"))

			child, err := childExpr.Exec(ctx, nil)
			if err != nil {
				return err
			}
			base, err := baseExpr.Exec(ctx, nil)
			if err != nil {
				return err
			}
			if child != "child-value" || base != "base-value" {
				return fmt.Errorf("unexpected values %v %v", child, base)
			}
			return nil
		}),
	)
}

func TestSelectAnyOfAndWhenBranches(t0 *testing.T) {
	Then(t0, "条件与选择表达式补齐未覆盖分支",
		ExpectMust(func() error {
			expr, err := From(Expression{
				"select",
				When(Pipe(Get("x"), Eq(1)), Eq(1)),
				"fallback",
			})
			if err != nil {
				return err
			}
			ctx := WithValueGetter(context.Background(), ValueGetterFunc(func(name string) (any, bool) {
				return 2, true
			}))
			out, err := expr.Exec(ctx, 0)
			if err != nil {
				return err
			}
			if out != "fallback" {
				return fmt.Errorf("unexpected select result %v", out)
			}
			return nil
		}),
		ExpectMust(func() error {
			expr, err := From(AnyOf(1, 2, 3))
			if err != nil {
				return err
			}
			out, err := expr.Exec(context.Background(), 2)
			if err != nil {
				return err
			}
			if out != true {
				return fmt.Errorf("unexpected anyOf result %v", out)
			}
			return nil
		}),
		ExpectDo(func() error {
			expr, err := From(When(Get("x"), Eq(1)))
			if err != nil {
				return err
			}
			_, err = expr.Exec(context.Background(), nil)
			return err
		}, ErrorMatch(mustRE("missing value getter"))),
	)
}

func mustRE(s string) *regexp.Regexp {
	return regexp.MustCompile(s)
}
