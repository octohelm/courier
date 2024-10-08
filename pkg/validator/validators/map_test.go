package validators

import (
	"errors"
	"testing"

	"github.com/davecgh/go-spew/spew"
	validatorerrors "github.com/octohelm/courier/pkg/validator/errors"
	"github.com/octohelm/courier/pkg/validator/internal"
	"github.com/octohelm/courier/pkg/validator/testutil"
	"github.com/octohelm/x/ptr"
	testingx "github.com/octohelm/x/testing"
)

func TestMapValidatorProvider(t *testing.T) {
	rules := [][2]string{
		{"@map[1,1000]", "@map[1,1000]"},
		{"@map<,@map[1,2]>[1,]", "@map<,@map[1,2]>[1,]"},
		{"@map<@string[0,],@map[1,2]>[1,]", "@map<@string[0,],@map[1,2]>[1,]"},
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

func TestMapValidator(t *testing.T) {
	t.Run("should be valid", func(t *testing.T) {
		cases := testutil.Cases{
			{
				Expect: []byte(`{"x":{"a":1}}`),
				Target: &struct {
					X map[string]int `json:"x" validate:"@map[1,]"`
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
					X map[string]int `json:"x" validate:"@map[1,]"`
				}{},
				Expect: []byte(`{"x":{}}`),
				ExpectError: func(err error, v any) bool {
					return errors.As(err, ptr.Ptr(&validatorerrors.ErrMissingRequired{}))
				},
			},
			{
				Input: []byte(`{"x":{"1":0}}`),
				Target: &struct {
					X map[string]int `json:"x" validate:"@map<@string[2],@int>"`
				}{},
				Expect: []byte(`{"x":{}}`),
				ExpectError: func(err error, v any) bool {
					spew.Dump(err)

					return errors.As(err, ptr.Ptr(&validatorerrors.OutOfRangeError{}))
				},
			},
		}

		testutil.Run(t, cases...)
	})
}
