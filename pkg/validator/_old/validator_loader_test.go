package _old

import (
	"context"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/octohelm/x/ptr"
	typesutil "github.com/octohelm/x/types"
	. "github.com/onsi/gomega"
)

func TestNewValidatorLoader(t *testing.T) {
	type SomeStruct struct {
		PtrString *string
		String    string
	}

	var val *string
	someStruct := &SomeStruct{}

	cases := []struct {
		valuesPass   []any
		valuesFailed []any
		rule         string
		typ          reflect.Type
		validator    *ValidatorLoader
	}{
		{
			[]any{
				reflect.ValueOf(someStruct).Elem().FieldByName("String"),
				"1",
			},
			[]any{"222"},
			"@string[1,2] = '1'",
			reflect.TypeOf(""),
			&ValidatorLoader{
				Optional:        true,
				DefaultValue:    []byte("1"),
				PreprocessStage: PreprocessSkip,
			},
		},
		{
			[]any{
				Duration(1 * time.Second),
				Duration(1 * time.Second),
			},
			[]any{},
			"@string",
			reflect.TypeOf(Duration(1 * time.Second)),
			&ValidatorLoader{
				PreprocessStage: PreprocessString,
			},
		},
		{
			[]any{
				val,
				reflect.ValueOf(someStruct).Elem().FieldByName("value"),
				reflect.ValueOf(val),
				ptr.Ptr("1"),
			},
			[]any{
				ptr.Ptr("222"),
			},
			"@string[1,2] = 2",
			reflect.TypeOf(ptr.Ptr("")),
			&ValidatorLoader{
				Optional:        true,
				DefaultValue:    []byte("2"),
				PreprocessStage: PreprocessPtr,
			},
		},
		{
			[]any{
				ptr.Ptr("1"),
				ptr.Ptr("22"),
			},
			[]any{
				ptr.Ptr(""),
				(*string)(nil),
			},
			"@string[1,2]",
			reflect.TypeOf(ptr.Ptr("")),
			&ValidatorLoader{
				PreprocessStage: PreprocessPtr,
			},
		},
	}

	for _, c := range cases {
		t.Run(fmt.Sprintf("%s %s", c.typ, c.rule), func(t *testing.T) {
			validator, err := Compile(context.Background(), []byte(c.rule), typesutil.FromRType(c.typ), nil)
			NewWithT(t).Expect(err).To(BeNil())
			if err != nil {
				return
			}

			loader := validator.(*ValidatorLoader)

			NewWithT(t).Expect(loader.Optional).To(Equal(c.validator.Optional))
			NewWithT(t).Expect(loader.PreprocessStage).To(Equal(c.validator.PreprocessStage))
			NewWithT(t).Expect(loader.DefaultValue).To(Equal(c.validator.DefaultValue))

			for _, v := range c.valuesPass {
				err := loader.Validate(v)
				NewWithT(t).Expect(err).To(BeNil())
			}

			for _, v := range c.valuesFailed {
				err := loader.Validate(v)
				NewWithT(t).Expect(err).NotTo(BeNil())
			}
		})
	}
}

func TestNewValidatorLoaderFailed(t *testing.T) {
	invalidRules := map[reflect.Type][]string{
		reflect.TypeOf(1): {
			"@string",
			"@int[1,2] = s",
		},
		reflect.TypeOf(""): {
			"@string<length, 1>",
			"@string[1,2] = 123",
		},
		reflect.TypeOf(Duration(1)): {
			"@string[,10] = 2ss",
		},
	}

	for typ := range invalidRules {
		for _, r := range invalidRules[typ] {
			t.Run(fmt.Sprintf("%s validate %s", typ, r), func(t *testing.T) {
				_, err := Compile(context.Background(), []byte(r), typesutil.FromRType(typ), nil)
				NewWithT(t).Expect(err).NotTo(BeNil())
				t.Log(err)
			})
		}
	}
}

type Duration time.Duration

func (d Duration) MarshalText() ([]byte, error) {
	return []byte(time.Duration(d).String()), nil
}

func (d *Duration) UnmarshalText(data []byte) error {
	dur, err := time.ParseDuration(string(data))
	if err != nil {
		return err
	}
	*d = Duration(dur)
	return nil
}
