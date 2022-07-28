package transformer

import (
	"context"

	"github.com/octohelm/courier/pkg/transformer/core"
	_ "github.com/octohelm/courier/pkg/transformer/html"
	_ "github.com/octohelm/courier/pkg/transformer/json"
	_ "github.com/octohelm/courier/pkg/transformer/multipart"
	_ "github.com/octohelm/courier/pkg/transformer/octet"
	_ "github.com/octohelm/courier/pkg/transformer/plain"
	_ "github.com/octohelm/courier/pkg/transformer/urlencoded"
	_ "github.com/octohelm/courier/pkg/transformer/xml"
	typesx "github.com/octohelm/x/types"
)

type Transformer = core.Transformer
type RequestParameter = core.RequestParameter
type Option = core.Option

func NewTransformer(ctx context.Context, tpe typesx.Type, opt Option) (Transformer, error) {
	return core.NewTransformer(ctx, tpe, opt)
}
