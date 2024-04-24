package core

import (
	"context"
	"reflect"

	"github.com/octohelm/courier/pkg/validator"
	typesx "github.com/octohelm/x/types"
)

type RequestParameter struct {
	Parameter
	TransformerOption Option
	Transformer       Transformer
	Validator         validator.Validator
}

func EachRequestParameter(ctx context.Context, tpe typesx.Type, each func(rp *RequestParameter)) error {
	errSet := validator.NewErrorSet()

	EachParameter(ctx, tpe, func(p *Parameter) bool {
		rp := &RequestParameter{}
		rp.Parameter = *p

		rp.TransformerOption.Name = rp.Name

		if flagTags, ok := rp.Tags["name"]; ok {
			rp.TransformerOption.Omitempty = flagTags.HasFlag("omitempty")
		}

		if mimeFlags, ok := rp.Tags["mime"]; ok {
			rp.TransformerOption.MIME = mimeFlags.Name()
			rp.TransformerOption.Strict = mimeFlags.HasFlag("strict")
			rp.TransformerOption.Omitempty = rp.TransformerOption.Strict
		}

		if rp.In == "path" {
			rp.TransformerOption.Omitempty = false
		}

		switch rp.Type.Kind() {
		case reflect.Array, reflect.Slice:
			if !isBytes(rp.Type) {
				rp.TransformerOption.Explode = true
			}
			_, ok := typesx.EncodingTextMarshalerTypeReplacer(rp.Type)
			if ok {
				rp.TransformerOption.Explode = false
			}
		default:

		}

		getTransformer := func() (Transformer, error) {
			if rp.TransformerOption.Explode {
				return NewTransformer(ctx, rp.Type.Elem(), rp.TransformerOption)
			}
			return NewTransformer(ctx, rp.Type, rp.TransformerOption)
		}

		tt, err := getTransformer()
		if err != nil {
			errSet.AddErr(err, rp.Name)
			return false
		}
		rp.Transformer = tt

		parameterValidator, err := NewValidator(ctx, rp.Type, rp.Tags, rp.TransformerOption.Omitempty, tt)
		if err != nil {
			errSet.AddErr(err, rp.Name)
			return false
		}
		rp.Validator = parameterValidator

		each(rp)

		return true
	})

	if errSet.Len() == 0 {
		return nil
	}

	return errSet.Err()
}

func isBytes(t typesx.Type) bool {
	return t.PkgPath() == "" && t.Kind() == reflect.Slice && t.Elem().PkgPath() == "" && t.Elem().Kind() == reflect.Uint8
}
