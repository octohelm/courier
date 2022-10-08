package raw

import (
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

func (arr ArrayValue) Iter() Iter {
	return &arrayValueIter{arr: arr, n: len(arr)}
}

type MapValue map[string]Value

func (MapValue) Kind() Kind {
	return Map
}

func (m MapValue) Keys() []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func (m MapValue) Iter() Iter {
	return &mapValueIter{keys: m.Keys(), m: m, n: len(m)}
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

type Iter interface {
	Next() bool
	Val() IterVal
}

type IterVal interface {
	Key() Value
	Value() Value
}

type arrayValueIter struct {
	arr ArrayValue
	n   int
	i   int
}

func (a *arrayValueIter) Next() bool {
	return a.i < a.n
}

func (a *arrayValueIter) Val() IterVal {
	v := &iterVal{
		value: a.arr.Index(a.i),
		key:   IntValue(a.i),
	}
	a.i++
	return v
}

type mapValueIter struct {
	keys []string
	m    MapValue
	n    int
	i    int
}

func (a *mapValueIter) Next() bool {
	return a.i < a.n
}

func (a *mapValueIter) Val() IterVal {
	v := &iterVal{
		value: a.m[a.keys[a.i]],
		key:   StringValue(a.keys[a.i]),
	}
	a.i++
	return v
}

type iterVal struct {
	key   Value
	value Value
}

func (i *iterVal) Key() Value {
	return i.key
}

func (i *iterVal) Value() Value {
	return i.value
}
