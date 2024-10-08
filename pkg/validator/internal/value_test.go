package internal_test

import (
	"testing"

	"github.com/go-json-experiment/json/jsontext"
	"github.com/octohelm/courier/pkg/validator/testutil"

	_ "github.com/octohelm/courier/pkg/validator/validators"
)

func TestValue(t *testing.T) {
	type SimpleStruct struct {
		Uint    int     `json:"uint,omitzero"`
		Int64   int     `json:"int64,omitzero"`
		Int8    int     `json:"int8,omitzero"`
		Float64 float64 `json:"float64,omitzero"`
		Float32 float32 `json:"float32,omitzero"`
		Bool    *bool   `json:"bool,omitzero"`
	}

	type WithEmbed struct {
		Map   map[string]string `json:"map,omitzero"`
		Slice []string          `json:"slice,omitzero"`

		*SimpleStruct
	}

	t.Run("unmarshal normal", func(t *testing.T) {
		cases := testutil.Cases{
			{
				Expect: []byte(`{"slice":["1","2"],"uint":1,"bool":false}`),
				Target: &WithEmbed{},
			},
			{
				Expect: []byte(`{"map":{"a":"a"},"uint":1,"bool":false}`),
				Target: &WithEmbed{},
			},
		}

		testutil.Run(t, cases...)
	})

	t.Run("unmarshal with default", func(t *testing.T) {
		cases := testutil.Cases{
			{
				Input:  []byte(`{}`),
				Expect: []byte(`{"x":"1"}`),
				Target: &struct {
					Value string `json:"x,omitzero" default:"1"`
				}{},
			},
			{
				Input:  []byte(`{}`),
				Expect: []byte(`{"x":1}`),
				Target: &struct {
					Value int `json:"x,omitzero" default:"1"`
				}{},
			},
		}

		testutil.Run(t, cases...)
	})

	t.Run("unmarshal any", func(t *testing.T) {
		cases := []testutil.Case{
			{
				Expect: []byte(`{"x":"1"}`),
				Target: &struct {
					Any any `json:"x,omitzero"`
				}{},
			},
			{
				Expect: []byte(`{}`),
				Target: &struct {
					Any any `json:"x,omitzero"`
				}{},
			},
		}

		testutil.Run(t, cases...)
	})

	t.Run("unmarshal unknown", func(t *testing.T) {
		cases := testutil.Cases{
			{
				Expect: []byte(`{"a":"a","x-a":"1"}`),
				Target: &struct {
					A       string         `json:"a"`
					Unknown jsontext.Value `json:",unknown"`
				}{},
			},
		}

		testutil.Run(t, cases...)
	})
}
