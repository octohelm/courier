package core

import (
	"github.com/octohelm/courier/pkg/transformer/internal"
)

func StringifyPath(path string, params map[string]string) string {
	return internal.NewPathnamePattern(path).Stringify(params)
}
