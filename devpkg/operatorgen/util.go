package operatorgen

import (
	"regexp"
)

// https://pkg.go.dev/fmt#hdr-Printing
var re = regexp.MustCompile(`%(\[([0-9]+)])?(([.0-9]+)|([#-+ 0]))?[vTtbcdoOqxXUeEfFgGps]`)

func normalizeFormat(s string) string {
	return re.ReplaceAllStringFunc(s, func(seg string) string {
		return "%v"
	})
}
