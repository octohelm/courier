package statuserror

import (
	"iter"
)

func All(err error) iter.Seq[error] {
	return func(yield func(error) bool) {
		if err == nil {
			return
		}

		if !(yield(err)) {
			return
		}

		switch x := err.(type) {
		case interface{ Unwrap() error }:
			err = x.Unwrap()
			if err == nil {
				return
			}
			if !(yield(err)) {
				return
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
		}
	}
}
