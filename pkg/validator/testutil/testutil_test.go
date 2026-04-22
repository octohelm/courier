package testutil

import (
	"testing"
)

import (
	_ "github.com/octohelm/courier/pkg/validator/validators"
)

func TestRun(t *testing.T) {
	t.Run("runs successful decode and expectation", func(t *testing.T) {
		Run(t, Case{
			Expect: []byte(`{"name":"demo"}`),
			Target: &struct {
				Name string `json:"name"`
			}{},
		})
	})

	t.Run("passes decode error to ExpectError callback", func(t *testing.T) {
		Run(t, Case{
			Input: []byte(`{"value":"abc"}`),
			Target: &struct {
				Value int `json:"value" validate:"@int"`
			}{},
			ExpectError: func(err error, v any) bool {
				return err != nil
			},
		})
	})
}
