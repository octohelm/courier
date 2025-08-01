package courier

import (
	"context"
	"fmt"
	"testing"

	"github.com/octohelm/x/testing/bdd"
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

func TestRegister(t *testing.T) {
	RouterRoot := NewRouter(&EmptyOperator{})
	RouterA := NewRouter(&OperatorA{})
	RouterRoot.Register(RouterA)

	bdd.FromT(t).When("register", func(b bdd.T) {
		b.Then("panic with conflict",
			bdd.HasError(Try(func() {
				RouterRoot.Register(RouterA)
			})),
		)
	})
}
