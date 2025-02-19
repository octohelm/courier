package jsonschema

import (
	"fmt"
	"os"
	"testing"

	"github.com/go-json-experiment/json"
	"github.com/octohelm/courier/internal/testingutil"
	testingx "github.com/octohelm/x/testing"
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

				testingx.Expect(t, err, testingx.BeNil[error]())
				testingx.Expect(t, c.expect(schema), testingx.Be(true))
			})
		}
	})

	t.Run("full", func(t *testing.T) {
		data, err := os.ReadFile("./testdata/2020-12/meta/applicator.json")
		testingx.Expect(t, err, testingx.Be[error](nil))

		p := &Payload{}
		err = p.UnmarshalJSON(data)
		testingx.Expect(t, err, testingx.Be[error](nil))

		p.Schema.PrintTo(os.Stdout)
	})
}

func TestSchema(t *testing.T) {
	t.Run("ref", func(t *testing.T) {
		testingx.Expect(t, Any(), testingutil.BeJSON[*AnyType](`{"x-go-type":"any"}`))
	})

	t.Run("any", func(t *testing.T) {
		testingx.Expect(t, Any(), testingutil.BeJSON[*AnyType](`{"x-go-type":"any"}`))
	})

	t.Run("string", func(t *testing.T) {
		testingx.Expect(t, String(), testingutil.BeJSON[*StringType](`{"type":"string"}`))
	})

	t.Run("bytes", func(t *testing.T) {
		testingx.Expect(t, Bytes(), testingutil.BeJSON[*StringType](`{"type":"string","format":"bytes"}`))
	})

	t.Run("binary", func(t *testing.T) {
		testingx.Expect(t, Binary(), testingutil.BeJSON[*StringType](`{"type":"string","format":"binary"}`))
	})

	t.Run("boolean", func(t *testing.T) {
		testingx.Expect(t, Boolean(), testingutil.BeJSON[*BooleanType](`{"type":"boolean"}`))
	})

	t.Run("array", func(t *testing.T) {
		testingx.Expect(t, ArrayOf(String()), testingutil.BeJSON[*ArrayType](`{"type":"array","items":{"type":"string"}}`))
	})

	t.Run("object", func(t *testing.T) {
		testingx.Expect(t,
			ObjectOf(
				map[string]Schema{
					"key1": String(),
					"key2": String(),
				},
				"key1",
			),
			testingutil.BeJSON[*ObjectType](`{"type":"object","properties":{"key1":{"type":"string"},"key2":{"type":"string"}},"required":["key1"]}`),
		)
	})

	t.Run("object with additional", func(t *testing.T) {
		testingx.Expect(t,
			MapOf(String()),
			testingutil.BeJSON[*ObjectType](`{"type":"object","additionalProperties":{"type":"string"}}`),
		)
	})

	t.Run("object with additionalProperties and propNames", func(t *testing.T) {
		testingx.Expect(t,
			RecordOf(String(), String()),
			testingutil.BeJSON[*ObjectType](`{"type":"object","propertyNames":{"type":"string"},"additionalProperties":{"type":"string"}}`),
		)
	})

	t.Run("oneOf", func(t *testing.T) {
		testingx.Expect(t,
			OneOf(String(), Boolean()),
			testingutil.BeJSON[*UnionType](`{"oneOf":[{"type":"string"},{"type":"boolean"}]}`),
		)
	})
}

func TestCommon(t *testing.T) {
	data, err := json.Marshal(Metadata{Ext: ExtOf(map[string]any{
		"x-v": "string",
	})})

	testingx.Expect(t, err, testingx.Be[error](nil))
	testingx.Expect(t, string(data), testingx.Be(`{"x-v":"string"}`))
}
