package raw

import (
	"fmt"
	"math"
)

// Mod
// v % x
func Mod(x Value, v Value) (any, error) {
	switch x.Kind() {
	case Float:
		switch v.Kind() {
		case Int, Uint, Float:
			return fixDecimal(math.Mod(ToFloat(v), ToFloat(x))), nil
		}
	case Int:
		switch v.Kind() {
		case Int, Uint:
			return ToInt(v) % ToInt(x), nil
		case Float:
			return fixDecimal(math.Mod(ToFloat(v), ToFloat(x))), nil
		}
	case Uint:
		switch v.Kind() {
		case Uint:
			return ToUint(v) % ToUint(x), nil
		case Int:
			return ToInt(v) % ToInt(x), nil
		case Float:
			return fixDecimal(math.Mod(ToFloat(v), ToFloat(x))), nil
		}
	}

	return nil, fmt.Errorf("%T can't mod %T", v, x)
}
