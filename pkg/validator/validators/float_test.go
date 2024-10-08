package validators

import (
	"errors"
	"testing"

	"github.com/go-json-experiment/json/jsontext"
	validatorerrors "github.com/octohelm/courier/pkg/validator/errors"
	"github.com/octohelm/courier/pkg/validator/internal"
	"github.com/octohelm/courier/pkg/validator/testutil"
	"github.com/octohelm/x/ptr"
	testingx "github.com/octohelm/x/testing"
)

func TestFloatValidatorProvider(t *testing.T) {
	rules := [][2]string{
		{"@float[1,1000]", "@float<7,2>[1,1000]"},
		{"@float32[1,1000]", "@float<7,2>[1,1000]"},
		{"@double[1,1000]", "@float<15,2>[1,1000]"},
		{"@float64[1,1000]", "@float<15,2>[1,1000]"},
		{"@float(1,1000]", "@float<7,2>(1,1000]"},
		{"@float[-1]", "@float<7,2>[-1,-1]"},
		{"@float[0,]", "@float<7,2>[0,]"},
		{"@float[0.1,]", "@float<7,2>[0.1,]"},
		{"@float{%2.2}", "@float<7,2>{%2.2}"},
		{"@float<10,3>[1.333,2.333]", "@float<10,3>[1.333,2.333]"},
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

func TestFloatValidator(t *testing.T) {
	t.Run("should be valid", func(t *testing.T) {
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

	t.Run("should be invalid", func(t *testing.T) {
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
					return errors.As(err, ptr.Ptr(&validatorerrors.OutOfRangeError{}))
				},
			},
			{
				Input: []byte(`{"x":-1}`),
				Target: &struct {
					X *float64 `json:"x,omitzero" validate:"@float[1,]"`
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

func Test_lengthOfDigits(t *testing.T) {
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
		n, d := lengthOfDigits(jsontext.Value(f.v))

		testingx.Expect(t, n, testingx.Be(f.n))
		testingx.Expect(t, d, testingx.Be(f.d))
	}
}
