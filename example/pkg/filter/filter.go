package filter

type Filter[X comparable] struct {
	v X
}

func (f *Filter[X]) UnmarshalText(data []byte) error {
	return nil
}

func (f Filter[X]) MarshalText() ([]byte, error) {
	return nil, nil
}
