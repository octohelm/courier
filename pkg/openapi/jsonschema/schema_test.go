package jsonschema

import (
	"fmt"
	"os"
	"testing"

	"github.com/go-json-experiment/json"
	. "github.com/octohelm/x/testing/v2"

	"github.com/octohelm/courier/internal/testingutil"
)

func TestSchemaUnmarshal(t *testing.T) {
	t.Run("normalize decoder", func(t *testing.T) {
		cases := []struct {
			desc   string
			data   string
			expect func(v any) bool
		}{
			{
				"true value should be any type",
				"true",
				IsType[*AnyType],
			},
			{
				"ref schema",
				`{ "$ref": "#/$defs/anchorString" }`,
				IsType[*RefType],
			},
			{
				"ref schema with allOf",
				`{ "allOf": [ { "$ref": "#/$defs/anchorString" }, { "description": "x" } ] }`,
				IsType[*RefType],
			},
			{
				"string schema",
				`{ "type": "string" }`,
				IsType[*StringType],
			},
			{
				"multiple types schema",
				`{ "type": ["object", "boolean"] }`,
				IsType[*UnionType],
			},
			{
				"const schema",
				`{ "const": "string" }`,
				IsType[*EnumType],
			},
			{
				"object schema",
				`{ "properties": { "a": { "type": "string" }, "b": { "type": "boolean" } } }`,
				IsType[*ObjectType],
			},
		}

		for _, c := range cases {
			t.Run(c.desc, func(t *testing.T) {
				var schema Schema

				err := json.Unmarshal([]byte(c.data), &schema, json.WithUnmarshalers(schemaUnmarshalers))
				if err != nil {
					fmt.Printf("#%v\n", err)
				}

				Then(t, "schema 会被解码为预期类型", ExpectMust(func() error {
					if err != nil {
						return err
					}
					if !c.expect(schema) {
						return fmt.Errorf("unexpected schema type for %s", c.desc)
					}
					return nil
				}))
			})
		}
	})

	t.Run("full", func(t *testing.T) {
		data, err := os.ReadFile("./testdata/2020-12/meta/applicator.json")
		Then(t, "可以读取完整 applicator schema", Expect(err, Equal[error](nil)))

		p := &Payload{}
		err = p.UnmarshalJSON(data)
		Then(t, "完整 schema 可成功反序列化", Expect(err, Equal[error](nil)))

		p.Schema.PrintTo(os.Stdout)
	})
}

func TestSchema(t *testing.T) {
	t.Run("ref", func(t *testing.T) {
		Then(t, "Any 会输出 any schema", Expect(Any(), Be(testingutil.BeJSON[*AnyType](`{"x-go-type":"any"}`))))
	})

	t.Run("any", func(t *testing.T) {
		Then(t, "Any helper 输出稳定 JSON", Expect(Any(), Be(testingutil.BeJSON[*AnyType](`{"x-go-type":"any"}`))))
	})

	t.Run("string", func(t *testing.T) {
		Then(t, "String helper 输出 string schema", Expect(String(), Be(testingutil.BeJSON[*StringType](`{"type":"string"}`))))
	})

	t.Run("bytes", func(t *testing.T) {
		Then(t, "Bytes helper 输出 bytes schema", Expect(Bytes(), Be(testingutil.BeJSON[*StringType](`{"type":"string","format":"bytes"}`))))
	})

	t.Run("binary", func(t *testing.T) {
		Then(t, "Binary helper 输出 binary schema", Expect(Binary(), Be(testingutil.BeJSON[*StringType](`{"type":"string","format":"binary"}`))))
	})

	t.Run("boolean", func(t *testing.T) {
		Then(t, "Boolean helper 输出 boolean schema", Expect(Boolean(), Be(testingutil.BeJSON[*BooleanType](`{"type":"boolean"}`))))
	})

	t.Run("array", func(t *testing.T) {
		Then(t, "ArrayOf helper 输出 array schema", Expect(ArrayOf(String()), Be(testingutil.BeJSON[*ArrayType](`{"type":"array","items":{"type":"string"}}`))))
	})

	t.Run("object", func(t *testing.T) {
		Then(t, "ObjectOf helper 输出 required properties", Expect(
			ObjectOf(
				map[string]Schema{
					"key1": String(),
					"key2": String(),
				},
				"key1",
			),
			Be(testingutil.BeJSON[*ObjectType](`{"type":"object","properties":{"key1":{"type":"string"},"key2":{"type":"string"}},"required":["key1"]}`)),
		))
	})

	t.Run("object with additional", func(t *testing.T) {
		Then(t, "MapOf helper 输出 additionalProperties", Expect(
			MapOf(String()),
			Be(testingutil.BeJSON[*ObjectType](`{"type":"object","additionalProperties":{"type":"string"}}`)),
		))
	})

	t.Run("object with additionalProperties and propNames", func(t *testing.T) {
		Then(t, "RecordOf helper 输出 propertyNames 和 additionalProperties", Expect(
			RecordOf(String(), String()),
			Be(testingutil.BeJSON[*ObjectType](`{"type":"object","propertyNames":{"type":"string"},"additionalProperties":{"type":"string"}}`)),
		))
	})

	t.Run("oneOf", func(t *testing.T) {
		Then(t, "OneOf helper 输出 union schema", Expect(
			OneOf(String(), Boolean()),
			Be(testingutil.BeJSON[*UnionType](`{"oneOf":[{"type":"string"},{"type":"boolean"}]}`)),
		))
	})
}

func TestCommon(t *testing.T) {
	data, err := json.Marshal(Metadata{Ext: ExtOf(map[string]any{
		"x-v": "string",
	})})

	Then(t, "Metadata 扩展字段可以被序列化",
		Expect(err, Equal[error](nil)),
		Expect(string(data), Equal(`{"x-v":"string"}`)),
	)
}
