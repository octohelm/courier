package jsonflags

import (
	"fmt"
	"iter"
	"reflect"
	"sync"

	"github.com/go-json-experiment/json"
)

type StructField struct {
	FieldOptions

	FieldName string
	Tag       reflect.StructTag
	Type      reflect.Type

	id    int
	index []int
}

func (f *StructField) GetOrNewAt(v reflect.Value) reflect.Value {
	if len(f.index) == 1 {
		return v.Field(f.index[0])
	}

	for i, x := range f.index {
		if i > 0 {
			if v.Kind() == reflect.Pointer {
				if v.IsNil() {
					v.Set(reflect.New(v.Type().Elem()))
				}
				v = v.Elem()
			}
		}
		v = v.Field(x)
	}

	return v

}

type StructFields struct {
	flattened       []*StructField
	byName          map[string]*StructField
	inlinedFallback *StructField
	located         map[string][]*StructField
	byLocatedName   map[locatedName]*StructField
}

type locatedName struct {
	location string
	name     string
}

func (s *StructFields) LookupLocated(location string, name string) (*StructField, bool) {
	sf, ok := s.byLocatedName[locatedName{location: location, name: name}]
	return sf, ok
}

func (s *StructFields) LocatedStructField(location string) iter.Seq[*StructField] {
	return func(yield func(*StructField) bool) {
		for _, sf := range s.located[location] {
			if !yield(sf) {
				return
			}
		}
	}
}

func (s *StructFields) StructField() iter.Seq[*StructField] {
	return func(yield func(*StructField) bool) {
		for _, sf := range s.flattened {
			if !yield(sf) {
				return
			}
		}
	}
}

func (s *StructFields) Lookup(name string) (*StructField, bool) {
	sf, ok := s.byName[name]
	return sf, ok
}

func (s *StructFields) InlinedFallback() (*StructField, bool) {
	return s.inlinedFallback, s.inlinedFallback != nil
}

func (s *StructFields) Len() int {
	return len(s.flattened)
}

var Structs = &cache{}

type cache struct {
	// map[reflect.Type]*StructFields
	structFields sync.Map
}

func (v *cache) StructFields(typ reflect.Type) (*StructFields, error) {
	if typ.Kind() == reflect.Ptr {
		panic(fmt.Errorf("invalid type %s", typ))
		return nil, nil
	}
	if vv, ok := v.structFields.Load(typ); ok {
		return vv.(*StructFields), nil
	}
	sfs, err := makeStructFields(typ)
	if err != nil {
		return nil, err
	}
	v.structFields.Store(typ, sfs)
	return sfs, nil
}

func makeStructFields(root reflect.Type) (*StructFields, *json.SemanticError) {
	type queueEntry struct {
		typ           reflect.Type
		index         []int
		visitChildren bool
	}

	seen := map[reflect.Type]bool{root: true}
	queue := []queueEntry{{root, nil, true}}
	queueIndex := 0

	allFields := make([]*StructField, 0)
	inlinedFallbacks := make([]*StructField, 0)

	for queueIndex < len(queue) {
		qe := queue[queueIndex]
		queueIndex++

		t := qe.typ
		for i := range t.NumField() {
			sf := t.Field(i)

			options, ignored, err := ParseFieldOptions(sf)
			if err != nil {
				return nil, &json.SemanticError{GoType: t, Err: err}
			} else if ignored {
				continue
			}

			f := &StructField{
				FieldName: sf.Name,
				Type:      sf.Type,
				Tag:       sf.Tag,

				FieldOptions: options,

				index: append(append(make([]int, 0, len(qe.index)+1), qe.index...), i),
			}

			if sf.Anonymous && !f.HasName {
				f.Inline = true
			}

			if f.Inline || f.Unknown {
				tf := f.Type

				if tf.Kind() == reflect.Pointer && tf.Name() == "" {
					tf = tf.Elem()
				}

				if tf.Kind() == reflect.Struct {
					if f.Unknown {
						continue
					}

					if qe.visitChildren {
						queue = append(queue, queueEntry{
							typ:           tf,
							index:         f.index,
							visitChildren: !seen[tf],
						})
					}

					seen[tf] = true

					continue
				}

				inlinedFallbacks = append(inlinedFallbacks, f)
			} else {
				f.id = len(allFields)
				allFields = append(allFields, f)
			}
		}
	}

	sfs := &StructFields{
		flattened:     allFields,
		byName:        make(map[string]*StructField, len(allFields)),
		byLocatedName: make(map[locatedName]*StructField, len(allFields)),
		located:       make(map[string][]*StructField, len(allFields)),
	}

	for _, f := range sfs.flattened {
		sfs.byName[f.Name] = f

		if location := f.Tag.Get("in"); location != "" {
			sfs.byLocatedName[locatedName{location: location, name: f.Name}] = f
			sfs.located[location] = append(sfs.located[location], f)
		}
	}

	if n := len(inlinedFallbacks); n == 1 || (n > 1 && len(inlinedFallbacks[0].index) != len(inlinedFallbacks[1].index)) {
		sfs.inlinedFallback = inlinedFallbacks[0]
	}

	return sfs, nil
}
