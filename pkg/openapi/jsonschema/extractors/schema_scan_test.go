package extractors

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"testing"
)

type schemaInvalidValidateType struct {
	Name string `json:"name" validate:"@["`
}

func TestSchemaErrorMessages(t *testing.T) {
	t.Run("非法 validate 规则应返回中文上下文", func(t *testing.T) {
		err := captureSchemaPanic(func() {
			_ = SchemaFromType(context.Background(), reflect.TypeFor[schemaInvalidValidateType](), Opt{})
		})
		if err == nil {
			t.Fatalf("expected panic error")
		}
		if !strings.Contains(err.Error(), `字段 Name 的 validate 规则 "@[" 非法`) {
			t.Fatalf("unexpected error message: %v", err)
		}
	})

	t.Run("非 string map key 应返回中文上下文", func(t *testing.T) {
		err := captureSchemaPanic(func() {
			_ = SchemaFromType(context.Background(), reflect.TypeFor[map[int]string](), Opt{})
		})
		if err == nil {
			t.Fatalf("expected panic error")
		}
		if !strings.Contains(err.Error(), "map key 仅支持 string schema") {
			t.Fatalf("unexpected error message: %v", err)
		}
	})

	t.Run("不支持的类型应返回中文上下文", func(t *testing.T) {
		err := captureSchemaPanic(func() {
			_ = SchemaFromType(context.Background(), reflect.TypeFor[func()](), Opt{})
		})
		if err == nil {
			t.Fatalf("expected panic error")
		}
		if !strings.Contains(err.Error(), "暂不支持生成 schema") {
			t.Fatalf("unexpected error message: %v", err)
		}
	})
}

func captureSchemaPanic(fn func()) (err error) {
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
