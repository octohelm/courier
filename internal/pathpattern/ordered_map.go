package pathpattern

import "iter"

type orderedMap[K comparable, V any] struct {
	m    map[K]V
	list []V
}

func (m *orderedMap[K, V]) Add(k K, v V) {
	if m.m == nil {
		m.m = map[K]V{}
	}
	if _, ok := m.m[k]; !ok {
		m.m[k] = v
		m.list = append(m.list, v)
	}
}

func (m *orderedMap[K, V]) Keys() iter.Seq[K] {
	return func(yield func(K) bool) {
		for k := range m.m {
			if !yield(k) {
				return
			}
		}
	}
}

func (m *orderedMap[K, V]) Values() iter.Seq[V] {
	return func(yield func(V) bool) {
		for _, item := range m.list {
			if !yield(item) {
				return
			}
		}
	}
}

func (m *orderedMap[K, V]) Len() int {
	return len(m.m)
}

func (m *orderedMap[K, V]) Get(k K) (V, bool) {
	v, ok := m.m[k]
	return v, ok
}
