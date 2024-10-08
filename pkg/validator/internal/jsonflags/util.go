package jsonflags

import "reflect"

func Implements(t reflect.Type, ifaceType reflect.Type) bool {
	return t.Implements(ifaceType) || reflect.PointerTo(t).Implements(ifaceType)
}
