package internal

import (
	"fmt"
	"io"
	"iter"

	"github.com/go-json-experiment/json"
	"github.com/go-json-experiment/json/jsontext"
	"github.com/octohelm/x/container/list"
)

type Pair[K comparable, V any] struct {
	Key   K
	Value V
}

type Record[K comparable, V any] struct {
	props   map[K]*list.Element[*Pair[K, V]]
	ll      list.List[*Pair[K, V]]
	created bool
}

func (r Record[K, V]) IsZero() bool {
	return len(r.props) == 0
}

func (r *Record[K, V]) Len() int {
	return len(r.props)
}

func (r *Record[K, V]) KeyValues() iter.Seq2[K, V] {
	return func(yield func(K, V) bool) {
		for el := r.ll.Front(); el != nil; el = el.Next() {
			if !yield(el.Value.Key, el.Value.Value) {
				return
			}
		}
	}
}

func (r *Record[K, V]) Get(key K) (V, bool) {
	if r.props != nil {
		v, ok := r.props[key]
		if ok {
			return v.Value.Value, true
		}
	}
	return *new(V), false
}

func (r *Record[K, V]) initOnce() {
	if !r.created {
		r.created = true
		r.props = map[K]*list.Element[*Pair[K, V]]{}
		r.ll.Init()
	}
}

func (r *Record[K, V]) Set(key K, value V) bool {
	r.initOnce()

	_, alreadyExist := r.props[key]
	if alreadyExist {
		r.props[key].Value.Value = value
		return false
	}

	element := &Pair[K, V]{Key: key, Value: value}
	r.props[key] = r.ll.PushBack(element)
	return true
}

func (r *Record[K, V]) Delete(key K) (didDelete bool) {
	if r.props == nil {
		return false
	}

	element, ok := r.props[key]
	if ok {
		r.ll.Remove(element)

		delete(r.props, key)
	}
	return ok
}

var _ json.UnmarshalerFrom = &Record[string, string]{}

func (r *Record[K, V]) UnmarshalJSONFrom(d *jsontext.Decoder) error {
	t, err := d.ReadToken()
	if err != nil {
		if err == io.EOF {
			return nil
		}
		return err
	}

	kind := t.Kind()

	if kind != '{' {
		return &json.SemanticError{
			JSONPointer: d.StackPointer(),
			Err:         fmt.Errorf("object should starts with `{`, but got `%s`", kind),
		}
	}

	if r == nil {
		*r = Record[K, V]{}
	}

	for kind := d.PeekKind(); kind != '}'; kind = d.PeekKind() {
		var key K
		if err := json.UnmarshalDecode(d, &key); err != nil {
			return err
		}

		var val V
		if err := json.UnmarshalDecode(d, &val); err != nil {
			return err
		}

		r.Set(key, val)
	}

	// read the close '}'
	if _, err := d.ReadToken(); err != nil {
		if err != io.EOF {
			return nil
		}
		return err
	}

	return nil
}

var _ json.MarshalerTo = Record[string, string]{}

func (r Record[K, V]) MarshalJSONTo(enc *jsontext.Encoder) error {
	if err := enc.WriteToken(jsontext.BeginObject); err != nil {
		return err
	}

	for k, v := range r.KeyValues() {
		if err := json.MarshalEncode(enc, k); err != nil {
			return err
		}

		if err := json.MarshalEncode(enc, v); err != nil {
			return err
		}
	}

	if err := enc.WriteToken(jsontext.EndObject); err != nil {
		return err
	}

	return nil
}
