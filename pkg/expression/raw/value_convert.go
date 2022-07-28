package raw

import "fmt"

func Len(value Value) int {
	switch x := value.(type) {
	case StringValue:
		return len(x)
	case ArrayValue:
		return len(x)
	case MapValue:
		return len(x)
	}
	return 0
}

func ToString(value Value) string {
	switch x := value.(type) {
	case StringValue:
		return string(x)
	case FloatValue:
		return fmt.Sprintf("%v", float64(x))
	case IntValue:
		return fmt.Sprintf("%d", x)
	case UintValue:
		return fmt.Sprintf("%d", x)
	case BoolValue:
		if x {
			return "true"
		}
		return "false"
	}
	return ""
}

func ToFloat(value Value) float64 {
	switch x := value.(type) {
	case FloatValue:
		return float64(x)
	case IntValue:
		return float64(x)
	case UintValue:
		return float64(x)
	case BoolValue:
		if x {
			return 1
		}
		return 0
	}
	return 0
}

func ToInt(value Value) int64 {
	switch x := value.(type) {
	case FloatValue:
		return int64(x)
	case IntValue:
		return int64(x)
	case UintValue:
		return int64(x)
	case BoolValue:
		if x {
			return 1
		}
		return 0
	}
	return 0
}

func ToUint(value Value) uint64 {
	switch x := value.(type) {
	case FloatValue:
		return uint64(x)
	case IntValue:
		return uint64(x)
	case UintValue:
		return uint64(x)
	case BoolValue:
		if x {
			return 1
		}
		return 0
	}
	return 0
}

func ToBool(value Value) bool {
	switch x := value.(type) {
	case FloatValue:
		return x != 0
	case IntValue:
		return x != 0
	case UintValue:
		return x != 0
	case BoolValue:
		return bool(x)
	}
	return false
}
