package extractors

import (
	"iter"
	"slices"
	"strings"

	"github.com/octohelm/courier/pkg/openapi/jsonschema"
)

func SetTitleOrDescription(metadata *jsonschema.Metadata, lines []string) {
	if metadata == nil {
		return
	}

	if len(lines) > 0 {
		metadata.Title = strings.TrimSpace(lines[0])

		if len(lines) > 1 {
			metadata.Description = strings.TrimSpace(strings.Join(slices.Collect(filterLine(slices.Values(lines[1:]))), "\n"))
		}
	}
}

func filterLine(seq iter.Seq[string]) iter.Seq[string] {
	return func(yield func(string) bool) {
		for l := range seq {
			if strings.HasPrefix(l, "openapi:") {
				continue
			}

			if !yield(l) {
				return
			}
		}
	}
}
