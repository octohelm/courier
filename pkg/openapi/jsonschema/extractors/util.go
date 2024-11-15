package extractors

import (
	"github.com/octohelm/courier/pkg/openapi/jsonschema"
	"strings"
)

func SetTitleOrDescription(metadata *jsonschema.Metadata, lines []string) {
	if metadata == nil {
		return
	}

	if len(lines) > 0 {
		metadata.Title = strings.TrimSpace(lines[0])

		if len(lines) > 1 {
			metadata.Description = strings.TrimSpace(strings.Join(lines[1:], "\n"))
		}
	}
}
