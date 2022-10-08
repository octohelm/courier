package raw

import (
	"fmt"
	"math"
)

// Pow
// v ** x
func Pow(x Value, v Value) (any, error) {
	switch x.Kind() {
	case Float:
		switch v.Kind() {
		case Int, Uint, Float:
			return fixDecimal(math.Pow(ToFloat(v), ToFloat(x))), nil
		}
	case Int:
		switch v.Kind() {
		case Int, Uint:
			return int64(math.Pow(ToFloat(v), ToFloat(x))), nil
		case Float:
			return fixDecimal(math.Pow(ToFloat(v), ToFloat(x))), nil
		}
	case Uint:
		switch v.Kind() {
		case Uint:
			return uint64(math.Pow(ToFloat(v), ToFloat(x))), nil
		case Int:
			return int64(math.Pow(ToFloat(v), ToFloat(x))), nil
		case Float:
			return fixDecimal(math.Pow(ToFloat(v), ToFloat(x))), nil
		}
	}

	return nil, fmt.Errorf("%T can't pow %T", v, x)
}
