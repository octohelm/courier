package pathpattern

import (
	"path"
	"strings"
)

func splitPath(p string) []string {
	p = cleanPath(p)
	if p[0] == '/' {
		p = p[1:]
	}
	if p == "" {
		return make([]string, 0)
	}
	return strings.Split(p, "/")
}

func cleanPath(p string) string {
	if p == "" {
		return "/"
	}
	if p[0] != '/' {
		p = "/" + p
	}
	np := path.Clean(p)
	if p[len(p)-1] == '/' && np != "/" {
		if len(p) == len(np)+1 && strings.HasPrefix(p, np) {
			np = p
		} else {
			np += "/"
		}
	}
	return np
}
