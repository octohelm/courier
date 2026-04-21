package courier

import (
	"context"
	"fmt"
	"iter"
	"net/url"
	"reflect"
)

// CanInit 表示可初始化的接口。
type CanInit interface {
	Init(ctx context.Context) error
}

// Operator 表示操作符接口，是courier框架的核心抽象。
type Operator interface {
	Output(ctx context.Context) (any, error)
}

// WithMiddleOperatorSeq 表示包含中间操作符序列的接口。
type WithMiddleOperatorSeq interface {
	MiddleOperators() iter.Seq[Operator]
}

type MiddleOperators []Operator

type WithMiddleOperators interface {
	MiddleOperators() MiddleOperators
}

type MetadataCarrier interface {
	Meta() Metadata
}

type OperatorWithParams interface {
	OperatorParams() map[string][]string
}

type OperatorWithoutOutput interface {
	Operator
	NoOutput()
}

type ContextProvider interface {
	Operator
	ContextKey() any
}

type DefaultsSetter interface {
	SetDefaults()
}

type OperatorInit interface {
	InitFrom(o Operator)
}

type OperatorNewer interface {
	New() Operator
}

func NewOperatorFactory(op Operator, last bool) *OperatorFactory {
	opType := typeOfOperator(reflect.TypeOf(op))
	if opType.Kind() != reflect.Struct {
		panic(fmt.Errorf("operator must be a struct type, got %#v", op))
	}

	meta := &OperatorFactory{}
	meta.IsLast = last
	meta.Operator = op

	if _, isOperatorWithoutOutput := op.(OperatorWithoutOutput); isOperatorWithoutOutput {
		meta.NoOutput = true
	}

	meta.Type = typeOfOperator(reflect.TypeOf(op))

	if operatorWithParams, ok := op.(OperatorWithParams); ok {
		meta.Params = operatorWithParams.OperatorParams()
	}

	if !meta.IsLast {
		if ctxKey, ok := op.(ContextProvider); ok {
			meta.ContextKey = ctxKey.ContextKey()
		} else {
			if ctxKey, ok := op.(oldContextProvider); ok {
				meta.ContextKey = ctxKey.ContextKey()
			} else {
				meta.ContextKey = meta.Type.String()
			}
		}
	}

	return meta
}

type oldContextProvider interface {
	ContextKey() string
}

func typeOfOperator(tpe reflect.Type) reflect.Type {
	for tpe.Kind() == reflect.Pointer {
		return typeOfOperator(tpe.Elem())
	}
	return tpe
}

// OperatorFactory 操作符工厂，用于创建和配置操作符。
type OperatorFactory struct {
	Type       reflect.Type
	ContextKey any
	NoOutput   bool
	Params     url.Values
	IsLast     bool
	Operator   Operator
}

func (o *OperatorFactory) String() string {
	s := ""
	if st, ok := o.Operator.(fmt.Stringer); ok {
		s = st.String()
	} else {
		s = o.Type.String()
	}

	if o.Params != nil {
		return s + "?" + o.Params.Encode()
	}

	return s
}

func (o *OperatorFactory) New() Operator {
	var op Operator

	if operatorNewer, ok := o.Operator.(OperatorNewer); ok {
		op = operatorNewer.New()
	} else {
		op = reflect.New(o.Type).Interface().(Operator)
	}

	if operatorInit, ok := op.(OperatorInit); ok {
		operatorInit.InitFrom(o.Operator)
	}

	if defaultsSetter, ok := op.(DefaultsSetter); ok {
		defaultsSetter.SetDefaults()
	}

	return op
}

// EmptyOperator 空操作符实现，用于基础操作符组合。
type EmptyOperator struct {
	OperatorWithoutOutput
}
