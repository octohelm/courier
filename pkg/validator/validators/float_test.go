package validators

import (
	"errors"
	"fmt"
	"testing"

	"github.com/go-json-experiment/json/jsontext"
	"github.com/octohelm/x/ptr"
	. "github.com/octohelm/x/testing/v2"

	validatorerrors "github.com/octohelm/courier/pkg/validator/errors"
	"github.com/octohelm/courier/pkg/validator/internal"
	"github.com/octohelm/courier/pkg/validator/testutil"
)

func TestFloatValidatorProvider(t *testing.T) {
	rules := [][2]string{
		{"@float[1,1000]", "@float[1,1000]"},
		{"@float32[1,1000]", "@float[1,1000]"},
		{"@double[1,1000]", "@float[1,1000]"},
		{"@float64[1,1000]", "@float[1,1000]"},
		{"@float(1,1000]", "@float(1,1000]"},
		{"@float[-1]", "@float[-1,-1]"},
		{"@float[0,]", "@float[0,]"},
		{"@float[0.1,]", "@float[0.1,]"},
		{"@float{%2.2}", "@float{%2.2}"},
		{"@float<10,3>[1.333,2.333]", "@float<10,3>[1.333,2.333]"},
	}

	for _, r := range rules {
		t.Run("parse "+r[0], func(t *testing.T) {
			Then(t, "float validator 规则会被规范化", ExpectMust(func() error {
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

func TestFloatValidator(t *testing.T) {
	t.Run("accepts valid float input", func(t *testing.T) {
		cases := testutil.Cases{
			{
				Expect: []byte(`{"x":1}`),
				Target: &struct {
					X float64 `json:"x" validate:"@float{1,2,3}"`
				}{},
			},
			{
				Expect: []byte(`{"x":3}`),
				Target: &struct {
					X float64 `json:"x" validate:"@float[2,4]"`
				}{},
			},
			{
				Expect: []byte(`{"x":-2.2}`),
				Target: &struct {
					X float64 `json:"x" validate:"@float{%2.2}"`
				}{},
			},
		}

		testutil.Run(t, cases...)
	})

	t.Run("rejects invalid float input", func(t *testing.T) {
		cases := testutil.Cases{
			{
				Input: []byte(`{}`),
				Target: &struct {
					X *float64 `json:"x" validate:"@float[1,]"`
				}{},
				Expect: []byte(`{"x":null}`),
				ExpectError: func(err error, v any) bool {
					return errors.As(err, ptr.Ptr(&validatorerrors.ErrMissingRequired{}))
				},
			},
			{
				Input: []byte(`{"x":1.0009}`),
				Target: &struct {
					X *float64 `json:"x,omitzero" validate:"@float<7,2>"`
				}{},
				Expect: []byte(`{}`),
				ExpectError: func(err error, v any) bool {
					return errors.As(err, ptr.Ptr(&validatorerrors.ErrOutOfRange{}))
				},
			},
			{
				Input: []byte(`{"x":-1}`),
				Target: &struct {
					X *float64 `json:"x,omitzero" validate:"@float[1,]"`
				}{},
				Expect: []byte(`{}`),
				ExpectError: func(err error, v any) bool {
					return errors.As(err, ptr.Ptr(&validatorerrors.ErrOutOfRange{}))
				},
			},
		}

		testutil.Run(t, cases...)
	})
}

func TestLengthOfDigits(t *testing.T) {
	floats := []struct {
		v string
		n uint
		d uint
	}{
		{"99999.99999", 10, 5},
		{"-0.19999999999999998", 17, 17},
		{"9223372036854775808", 19, 0},
		{"340282346638528859811704183484516925440", 39, 0},
	}

	for _, f := range floats {
		t.Run(f.v, func(t *testing.T) {
			Then(t, "会返回整数位和小数位长度", ExpectMust(func() error {
				n, d := lengthOfDigits(jsontext.Value(f.v))
				if n != f.n || d != f.d {
					return fmt.Errorf("unexpected digits: n=%d d=%d", n, d)
				}
				return nil
			}))
		})
	}
}
