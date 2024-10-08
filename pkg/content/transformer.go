package content

import (
	"reflect"

	"github.com/octohelm/courier/pkg/content/internal"

	_ "github.com/octohelm/courier/pkg/content/transformers"
)

type Transformer = internal.Transformer

func New(typ reflect.Type, mediaTypeOrAlias string, action string) (Transformer, error) {
	return internal.New(typ, mediaTypeOrAlias, action)
}
