package courier

import (
	"context"
	"fmt"
	"strings"
	"testing"

	. "github.com/octohelm/x/testing/v2"
)

func ExampleNewRouter() {
	RouterRoot := NewRouter(&EmptyOperator{})
	RouterA := NewRouter(&OperatorA{})
	RouterB := NewRouter(&OperatorB{})

	RouterRoot.Register(RouterA)
	RouterRoot.Register(RouterB)

	RouterA.Register(NewRouter(&OperatorA1{}))
	RouterA.Register(NewRouter(&OperatorA2{}))
	RouterB.Register(NewRouter(&OperatorB2{}))

	fmt.Println(RouterRoot.Routes())
	// Output:
	// courier.EmptyOperator |> courier.OperatorA |> courier.OperatorA1?allowedRoles=ADMIN&allowedRoles=OWNER
	// courier.EmptyOperator |> courier.OperatorA |> courier.OperatorA2
	// courier.EmptyOperator |> courier.OperatorB |> courier.OperatorB1 |> courier.OperatorB2
}

type OperatorA struct{}

func (OperatorA) ContextKey() string {
	return "OperatorA"
}

func (OperatorA) Output(ctx context.Context) (any, error) {
	return nil, nil
}

type OperatorA1 struct{}

func (OperatorA1) OperatorParams() map[string][]string {
	return map[string][]string{
		"allowedRoles": {"ADMIN", "OWNER"},
	}
}

func (OperatorA1) Output(ctx context.Context) (any, error) {
	return nil, nil
}

type OperatorA2 struct{}

func (OperatorA2) Output(ctx context.Context) (any, error) {
	return nil, nil
}

type OperatorB struct{}

func (OperatorB) Output(ctx context.Context) (any, error) {
	return nil, nil
}

type OperatorB1 struct{}

func (OperatorB1) Output(ctx context.Context) (any, error) {
	return nil, nil
}

type OperatorB2 struct{}

func (OperatorB2) MiddleOperators() MiddleOperators {
	return MiddleOperators{
		&OperatorB1{},
	}
}

func (OperatorB2) Output(ctx context.Context) (any, error) {
	return nil, nil
}

func TestRegisterRejectsDuplicateRouter(t *testing.T) {
	RouterRoot := NewRouter(&EmptyOperator{})
	RouterA := NewRouter(&OperatorA{})
	RouterRoot.Register(RouterA)

	Then(t, "重复注册同一路由会触发冲突 panic",
		ExpectMust(func() error {
			err := Try(func() {
				RouterRoot.Register(RouterA)
			})
			if err == nil {
				return fmt.Errorf("expected register conflict")
			}
			return nil
		}),
	)
}

func TestRegisterConflictMessage(t *testing.T) {
	RouterRoot := NewRouter(&EmptyOperator{})
	RouterA := NewRouter(&OperatorA{})
	RouterRoot.Register(RouterA)

	Then(t, "冲突消息会包含重复注册与父路由上下文",
		ExpectMust(func() error {
			err := captureRouterPanic(func() {
				RouterRoot.Register(RouterA)
			})
			if err == nil {
				return fmt.Errorf("expected panic error")
			}
			if !strings.Contains(err.Error(), "路由重复注册") {
				return fmt.Errorf("unexpected error message: %v", err)
			}
			if !strings.Contains(err.Error(), "当前父路由") {
				return fmt.Errorf("missing parent route context: %v", err)
			}
			return nil
		}),
	)
}

func captureRouterPanic(fn func()) (err error) {
	defer func() {
		if x := recover(); x != nil {
			if e, ok := x.(error); ok {
				err = e
				return
			}
			err = fmt.Errorf("unexpected non-error panic: %v", x)
		}
	}()

	fn()
	return nil
}
