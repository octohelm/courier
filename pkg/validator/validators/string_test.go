package validators

import (
	"errors"
	"testing"

	validatorerrors "github.com/octohelm/courier/pkg/validator/errors"
	"github.com/octohelm/courier/pkg/validator/internal"
	"github.com/octohelm/courier/pkg/validator/testutil"
	"github.com/octohelm/x/ptr"
	testingx "github.com/octohelm/x/testing"
)

func TestStringValidatorProvider(t *testing.T) {
	rules := [][2]string{
		{"@string[1,1000]", "@string<length>[1,1000]"},
		{"@string[1,]", "@string<length>[1,]"},
		{"@string<length>[1]", "@string<length>[1,1]"},
		{"@char[1,]", "@string<rune_count>[1,]"},
		{"@string{KEY1,KEY2}", "@string<length>{KEY1,KEY2}"},
		{"@string/^\\w+/", "@string<length>/^\\w+/"},
		{"@string/^\\w+\\/test/", "@string<length>/^\\w+\\/test/"},
	}

	for _, r := range rules {
		t.Run("parse "+r[0], func(t *testing.T) {
			v, err := internal.New(internal.ValidatorOption{
				Rule: r[0],
			})
			testingx.Expect(t, err, testingx.BeNil[error]())
			testingx.Expect(t, v.String(), testingx.Be(r[1]))
		})
	}
}

func TestStringValidator(t *testing.T) {
	t.Run("should be valid", func(t *testing.T) {
		cases := testutil.Cases{
			{
				Expect: []byte(`{"x":"1"}`),
				Target: &struct {
					Int string `json:"x" validate:"@string[1,]"`
				}{},
			},
			{
				Expect: []byte(`{"x":"word"}`),
				Target: &struct {
					Int string `json:"x" validate:"@string/^\\w+/"`
				}{},
			},
		}

		testutil.Run(t, cases...)
	})

	t.Run("should be invalid", func(t *testing.T) {
		cases := testutil.Cases{
			{
				Input: []byte(`{}`),
				Target: &struct {
					Int *string `json:"x" validate:"@string[1,]"`
				}{},
				Expect: []byte(`{"x":null}`),
				ExpectError: func(err error, v any) bool {
					return errors.As(err, ptr.Ptr(&validatorerrors.ErrMissingRequired{}))
				},
			},
			{
				Input: []byte(`{"x":""}`),
				Target: &struct {
					Int *string `json:"x,omitzero" validate:"@string[1,]"`
				}{},
				Expect: []byte(`{}`),
				ExpectError: func(err error, v any) bool {
					return errors.As(err, ptr.Ptr(&validatorerrors.ErrOutOfRange{}))
				},
			},
			{
				Input: []byte(`{"x":"-word"}`),
				Target: &struct {
					Int *string `json:"x,omitzero" validate:"@string/^\\w+/"`
				}{},
				ExpectError: func(err error, v any) bool {
					return errors.As(err, ptr.Ptr(&validatorerrors.ErrPatternNotMatch{}))
				},
			},
		}

		testutil.Run(t, cases...)
	})
}
