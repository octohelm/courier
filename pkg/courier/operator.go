package courier

import (
	"context"
	"fmt"
	"iter"
	"net/url"
	"reflect"
)

type CanInit interface {
	Init(ctx context.Context) error
}

type Operator interface {
	Output(ctx context.Context) (any, error)
}

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
	for tpe.Kind() == reflect.Ptr {
		return typeOfOperator(tpe.Elem())
	}
	return tpe
}

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

type EmptyOperator struct {
	OperatorWithoutOutput
}
