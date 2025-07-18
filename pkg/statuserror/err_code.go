package statuserror

import (
	"go/ast"
	"path"
	"reflect"
	"unicode"
)

const (
	ERR_CODD_INVALID_PARAMETER     = "INVALID_PARAMETER"
	ERR_CODD_INTERNAL_SERVER_ERROR = "INTERNAL_SERVER_ERROR"
)

func ErrCodeFor[E any]() string {
	return ErrCodeOf(new(E))
}

func ErrCodeOf(err any) string {
	rv := reflect.TypeOf(err)
	for rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}

	if ast.IsExported(rv.Name()) {
		if pkgPath := rv.PkgPath(); pkgPath != "" {
			dir, p := path.Split(pkgPath)
			if isVersionSegment(p) {
				p = path.Base(dir) + p
			}
			return p + "." + rv.Name()
		}
		return rv.Name()
	}
	return ""
}

func isVersionSegment(s string) bool {
	if len(s) >= 2 {
		// v{number}...
		return s[0] == 'v' && unicode.IsDigit(rune(s[1]))
	}
	return false
}
