package core

import (
	"github.com/octohelm/courier/internal/pathpattern"
)

func StringifyPath(path string, params map[string]string) string {
	return pathpattern.Parse(path).Encode(params)
}
