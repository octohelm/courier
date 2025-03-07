package jsonflags

func ValueWithStructField(v any, sf *StructField) any {
	return &wrapValue{
		v:  v,
		sf: sf,
	}
}

type wrapValue struct {
	v  any
	sf *StructField
}

func (f *wrapValue) Unwrap() any {
	return f.v
}

func (f *wrapValue) StructField() *StructField {
	return f.sf
}

type Wrapper interface {
	Unwrap() any
	StructField() *StructField
}

func Unwrap(v any) any {
	if x, ok := v.(Wrapper); ok {
		return x.Unwrap()
	}
	return v
}
