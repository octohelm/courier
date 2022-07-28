package raw

import (
	"fmt"

	"github.com/pkg/errors"
)

// Div
// v / x
func Div(x Value, v Value) (any, error) {
	switch x.Kind() {
	case Float:
		if ToFloat(x) == 0 {
			return nil, errors.New("can't divide 0")
		}
		switch v.Kind() {
		case Int, Uint, Float:
			return ToFloat(v) / ToFloat(x), nil
		}
	case Int:
		if ToInt(x) == 0 {
			return nil, errors.New("can't divide 0")
		}

		switch v.Kind() {
		case Int, Uint:
			if ToInt(v)%ToInt(x) == 0 {
				return ToInt(v) / ToInt(x), nil
			}
			return ToFloat(v) / ToFloat(x), nil
		case Float:
			return ToFloat(v) / ToFloat(x), nil
		}
	case Uint:
		if ToUint(x) == 0 {
			return nil, errors.New("can't divide 0")
		}
		switch v.Kind() {
		case Uint:
			if ToUint(v)%ToUint(x) == 0 {
				return ToUint(v) / ToUint(x), nil
			}
			return ToFloat(v) / ToFloat(x), nil
		case Int:
			if ToInt(v)%ToInt(x) == 0 {
				return ToInt(v) / ToInt(x), nil
			}
			return ToFloat(v) / ToFloat(x), nil
		case Float:
			return ToFloat(v) / ToFloat(x), nil
		}
	}

	return nil, fmt.Errorf("%T can't divide %T", v, x)
}
