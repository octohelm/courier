package validators

import (
	"errors"
	"fmt"
	"testing"

	"github.com/octohelm/x/ptr"
	. "github.com/octohelm/x/testing/v2"

	validatorerrors "github.com/octohelm/courier/pkg/validator/errors"
	"github.com/octohelm/courier/pkg/validator/internal"
	"github.com/octohelm/courier/pkg/validator/testutil"
)

func TestSliceValidatorProvider(t *testing.T) {
	rules := [][2]string{
		{"@slice[1,1000]", "@slice[1,1000]"},
		{"@slice<@string[1,2]>[1,]", "@slice<@string[1,2]>[1,]"},
		{"@slice[1]", "@slice[1,1]"},
	}

	for _, r := range rules {
		t.Run("parse "+r[0], func(t *testing.T) {
			Then(t, "slice validator 规则会被规范化", ExpectMust(func() error {
				v, err := internal.New(internal.ValidatorOption{Rule: r[0]})
				if err != nil {
					return err
				}
				if v.String() != r[1] {
					return fmt.Errorf("unexpected validator string: %s", v.String())
				}
				return nil
			}))
		})
	}
}

func TestSliceValidator(t *testing.T) {
	t.Run("accepts valid slice input", func(t *testing.T) {
		cases := testutil.Cases{
			{
				Expect: []byte(`{"x":["1"]}`),
				Target: &struct {
					X []string `json:"x" validate:"@slice[1,]"`
				}{},
			},
		}

		testutil.Run(t, cases...)
	})

	t.Run("rejects invalid slice input", func(t *testing.T) {
		cases := testutil.Cases{
			{
				Input: []byte(`{}`),
				Target: &struct {
					X []int `json:"x" validate:"@slice[1,]"`
				}{},
				Expect: []byte(`{"x":[]}`),
				ExpectError: func(err error, v any) bool {
					return errors.As(err, ptr.Ptr(&validatorerrors.ErrMissingRequired{}))
				},
			},
			{
				Input: []byte(`{"x":[1,2,3]}`),
				Target: &struct {
					X []int `json:"x,omitzero" validate:"@slice[2]"`
				}{},
				Expect: []byte(`{"x":[1,2,3]}`),
				ExpectError: func(err error, v any) bool {
					return errors.As(err, ptr.Ptr(&validatorerrors.ErrOutOfRange{}))
				},
			},
			{
				Input: []byte(`{"x":[1,2,3]}`),
				Target: &struct {
					X []int `json:"x,omitzero" validate:"@slice<@int[1]>"`
				}{},
				Expect: []byte(`{"x":[1,0,0]}`),
				ExpectError: func(err error, v any) bool {
					return errors.As(err, ptr.Ptr(&validatorerrors.ErrOutOfRange{}))
				},
			},
		}

		testutil.Run(t, cases...)
	})
}
