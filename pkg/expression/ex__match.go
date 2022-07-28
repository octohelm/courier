package expression

import (
	"context"
	"regexp"

	"github.com/octohelm/courier/pkg/expression/raw"
)

// Match
//
// Syntax
//
//	match("[a-z]+")
var Match = Register(func(b ExprBuilder) func(pattern string) Expression {
	return func(pattern string) Expression {
		return b.BuildExpression(pattern)
	}
}, &match{})

type match struct {
	E
	Pattern regexpString
}

func (e *match) Exec(ctx context.Context, in any) (any, error) {
	return e.Pattern.MatchString(raw.ToString(raw.ValueOf(in))), nil
}

type regexpString struct {
	regexp.Regexp
}

func (r *regexpString) UnmarshalText(text []byte) error {
	re, err := regexp.Compile(string(text))
	if err != nil {
		return err
	}
	r.Regexp = *re
	return nil
}
