package extractors

import (
	"fmt"
	"strings"
	"testing"

	"github.com/octohelm/courier/pkg/openapi/jsonschema"
	"github.com/octohelm/courier/pkg/validator"
)

func TestPatchSchemaValidationErrorMessages(t *testing.T) {
	t.Run("slice 元素校验失败应返回中文上下文", func(t *testing.T) {
		err := captureValidatorPanic(func() {
			_, _ = PatchSchemaValidation(
				jsonschema.ArrayOf(jsonschema.String()),
				validator.Option{
					Rule: "@slice<@unknown>[1]",
				},
			)
		})
		if err == nil {
			t.Fatalf("expected panic error")
		}
		if !strings.Contains(err.Error(), "补充 slice 元素校验失败") {
			t.Fatalf("unexpected error message: %v", err)
		}
	})

	t.Run("map key 校验失败应返回中文上下文", func(t *testing.T) {
		err := captureValidatorPanic(func() {
			_, _ = PatchSchemaValidation(
				jsonschema.RecordOf(jsonschema.String(), jsonschema.String()),
				validator.Option{
					Rule: "@map<@unknown,@string>",
				},
			)
		})
		if err == nil {
			t.Fatalf("expected panic error")
		}
		if !strings.Contains(err.Error(), "补充 map key 校验失败") {
			t.Fatalf("unexpected error message: %v", err)
		}
	})

	t.Run("map value 校验失败应返回中文上下文", func(t *testing.T) {
		err := captureValidatorPanic(func() {
			_, _ = PatchSchemaValidation(
				jsonschema.RecordOf(jsonschema.String(), jsonschema.String()),
				validator.Option{
					Rule: "@map<@string,@unknown>",
				},
			)
		})
		if err == nil {
			t.Fatalf("expected panic error")
		}
		if !strings.Contains(err.Error(), "补充 map value 校验失败") {
			t.Fatalf("unexpected error message: %v", err)
		}
	})
}

func captureValidatorPanic(fn func()) (err error) {
	defer func() {
		if x := recover(); x != nil {
			switch e := x.(type) {
			case error:
				err = e
			default:
				err = fmt.Errorf("unexpected non-error panic: %v", x)
			}
		}
	}()

	fn()
	return nil
}
