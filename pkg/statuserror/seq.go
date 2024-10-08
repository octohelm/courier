package statuserror

import (
	"iter"
)

func All(err error) iter.Seq[error] {
	return func(yield func(error) bool) {
		if err == nil {
			return
		}

		switch x := err.(type) {
		case WithJSONPointer:
			if !(yield(err)) {
				return
			}
		case WithStatusCode:
			if !(yield(err)) {
				return
			}
		case interface{ Unwrap() error }:
			if _, ok := err.(WithLocation); ok {
				if !(yield(err)) {
					return
				}
			}

			err = x.Unwrap()
			if err == nil {
				return
			}

			for e := range All(err) {
				if !(yield(e)) {
					return
				}
			}
		case interface{ Unwrap() []error }:
			for _, ee := range x.Unwrap() {
				if ee == nil {
					continue
				}
				for e := range All(ee) {
					if !(yield(e)) {
						return
					}
				}
			}
		default:
			if !(yield(err)) {
				return
			}
		}
	}
}
