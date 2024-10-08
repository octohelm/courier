package validators

import (
	"errors"
	validatorerrors "github.com/octohelm/courier/pkg/validator/errors"
	"github.com/octohelm/x/ptr"
	"testing"

	"github.com/octohelm/courier/pkg/validator/testutil"
)

func TestInteger(t *testing.T) {
	t.Run("should be valid", func(t *testing.T) {
		cases := testutil.Cases{
			{
				Expect: []byte(`{"x":1}`),
				Target: &struct {
					Int int8 `json:"x" validate:"@int8[1,]"`
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
					Int *int8 `json:"x" validate:"@int8[1,]"`
				}{},
				Expect: []byte(`{"x":null}`),
				ExpectError: func(err error, v any) bool {
					return errors.As(err, ptr.Ptr(&validatorerrors.MissingRequired{}))
				},
			},
			{
				Input: []byte(`{"x":-1}`),
				Target: &struct {
					Int *int8 `json:"x,omitzero" validate:"@int8[1,]"`
				}{},
				Expect: []byte(`{}`),
				ExpectError: func(err error, v any) bool {
					return errors.As(err, ptr.Ptr(&validatorerrors.OutOfRangeError{}))
				},
			},
		}

		testutil.Run(t, cases...)
	})
}
