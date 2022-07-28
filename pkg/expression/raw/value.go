package raw

import (
	"context"
	"encoding"
)

func ValueOf(in any) Value {
	switch v := in.(type) {
	case Value:
		return v
	case encoding.TextMarshaler:
		b, _ := v.MarshalText()
		return StringValue(b)
	case float64:
		return FloatValue(v)
	case float32:
		return FloatValue(v)
	case int:
		return IntValue(v)
	case int8:
		return IntValue(v)
	case int16:
		return IntValue(v)
	case int32:
		return IntValue(v)
	case int64:
		return IntValue(v)
	case uint:
		return UintValue(v)
	case uint8:
		return UintValue(v)
	case uint16:
		return UintValue(v)
	case uint32:
		return UintValue(v)
	case uint64:
		return UintValue(v)
	case string:
		return StringValue(v)
	case bool:
		return BoolValue(v)
	case []any:
		arr := make(ArrayValue, len(v))
		for i := range v {
			arr[i] = ValueOf(v[i])
		}
		return arr
	case map[string]any:
		m := MapValue{}
		for k := range v {
			m[k] = ValueOf(v[k])
		}
		return m
	}
	return nil
}

type Kind int

const (
	Invalid Kind = iota
	Uint
	Int
	Float
	String
	Bool
	Array
	Map
)

type Value interface {
	Kind() Kind
}

type ArrayValue []Value

func (ArrayValue) Kind() Kind {
	return Array
}

func (arr ArrayValue) Len() (i int) {
	return len(arr)
}

func (arr ArrayValue) Index(i int) Value {
	return arr[i]
}

func (arr ArrayValue) Iter(ctx context.Context) <-chan Entity {
	ch := make(chan Entity, 1)
	go func() {
		defer close(ch)
		for i := range arr {
			e := &entity{
				key:   IntValue(i),
				value: arr.Index(i),
			}
			select {
			case <-ctx.Done():
				return
			case ch <- e:
			}
		}
	}()
	return ch
}

type MapValue map[string]Value

func (MapValue) Kind() Kind {
	return Map
}

func (m MapValue) Iter(ctx context.Context) <-chan Entity {
	ch := make(chan Entity, 1)

	go func() {
		defer close(ch)

		for k := range m {
			e := &entity{
				key:   StringValue(k),
				value: m[k],
			}

			select {
			case <-ctx.Done():
				return
			case ch <- e:
			}
		}
	}()

	return ch
}

type FloatValue float64

func (FloatValue) Kind() Kind {
	return Float
}

type IntValue int64

func (IntValue) Kind() Kind {
	return Int
}

type UintValue uint64

func (UintValue) Kind() Kind {
	return Uint
}

type StringValue string

func (StringValue) Kind() Kind {
	return String
}

type BoolValue bool

func (BoolValue) Kind() Kind {
	return Bool
}

type Iterable interface {
	Iter(ctx context.Context) <-chan Entity
}

type Entity interface {
	Key() Value
	Value() Value
}

type entity struct {
	key   Value
	value Value
}

func (i *entity) Key() Value {
	return i.key
}

func (i *entity) Value() Value {
	return i.value
}
