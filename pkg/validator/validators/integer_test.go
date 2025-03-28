package validators

import (
	"errors"
	"testing"

	validatorerrors "github.com/octohelm/courier/pkg/validator/errors"
	"github.com/octohelm/courier/pkg/validator/internal"
	"github.com/octohelm/x/ptr"
	testingx "github.com/octohelm/x/testing"

	"github.com/octohelm/courier/pkg/validator/testutil"
)

func TestIntegerValidatorProvider(t *testing.T) {
	rules := [][2]string{
		{"@uint8", "@uint<8>[0,255]"},
		{"@uint8[1,]", "@uint<8>[1,255]"},
		{"@int8", "@int<8>[-128,127]"},
		{"@uint16", "@uint<16>[0,65535]"},
		{"@int16", "@int<16>[-32768,32767]"},
		{"@int64", "@int<64>[-9223372036854775808,9223372036854775807]"},
		{"@int[1,1000)", "@int<32>[1,1000)"},
		{"@int(1,1000]", "@int<32>(1,1000]"},
		{"@uint16{1,2}", "@uint<16>{1,2}"},
		{"@uint16{%2}", "@uint<16>[0,65535]{%2}"},
		{"@int<53>", "@int<53>[-4503599627370496,4503599627370495]"},
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

func TestIntegerValidator(t *testing.T) {
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
					return errors.As(err, ptr.Ptr(&validatorerrors.ErrMissingRequired{}))
				},
			},
			{
				Input: []byte(`{"x":100000}`),
				Target: &struct {
					Int *uint `json:"x,omitzero" validate:"@uint8"`
				}{},
				Expect: []byte(`{}`),
				ExpectError: func(err error, v any) bool {
					return errors.As(err, ptr.Ptr(&validatorerrors.ErrOutOfRange{}))
				},
			},
			{
				Input: []byte(`{"x":-1}`),
				Target: &struct {
					Int *int8 `json:"x,omitzero" validate:"@int8[1,]"`
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
