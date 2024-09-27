package httprouter

import (
	"cmp"
	"strings"

	"github.com/octohelm/courier/internal/pathpattern"
)

func toHttpRouterPathPrefix(pathSegments pathpattern.Segments) string {
	s := &strings.Builder{}

	segN := len(pathSegments)

	for i, seg := range pathSegments {
		switch x := seg.(type) {
		case pathpattern.NamedSegment:
			s.WriteString("/")
			if x.Multiple() && i == (segN-1) {
				s.WriteString(x.String())
			} else {
				s.WriteString("{")
				s.WriteString(x.Name())
				s.WriteString("}")
			}
		default:
			s.WriteString("/")
			s.WriteString(x.String())
		}
	}

	return cmp.Or(s.String(), "/")
}
