package core

import (
	"context"

	"github.com/octohelm/courier/pkg/validator"
	typex "github.com/octohelm/x/types"
)

type MayValidator interface {
	NewValidator(ctx context.Context, typ typex.Type) (validator.Validator, error)
}

type WithNamedByTag interface {
	NamedByTag() string
}

func NewValidator(ctx context.Context, fieldType typex.Type, tags map[string]Tag, omitempty bool, transformer Transformer) (validator.Validator, error) {
	if withNamedByTag, ok := transformer.(WithNamedByTag); ok {
		if namedTagKey := withNamedByTag.NamedByTag(); namedTagKey != "" {
			ctx = validator.ContextWithNamedTagKey(ctx, namedTagKey)
		}
	}

	if t, ok := transformer.(MayValidator); ok {
		return t.NewValidator(ctx, fieldType)
	}

	mgr := validator.MgrFromContext(ctx)

	tagValidate := tags["validate"]

	return mgr.Compile(ctx, []byte(tagValidate), fieldType, func(rule validator.RuleModifier) {
		if omitempty {
			rule.SetOptional(true)
		}
		if defaultValue, ok := tags["default"]; ok {
			rule.SetDefaultValue([]byte(defaultValue))
		}
	})
}
