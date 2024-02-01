package jsonschema

import (
	"fmt"
	"os"
	"testing"

	"github.com/go-json-experiment/json"

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

				testingutil.Expect(t, err, testingutil.Be[error](nil))
				testingutil.Expect(t, c.expect(schema), testingutil.Be(true))
			})
		}
	})

	t.Run("full", func(t *testing.T) {
		data, err := os.ReadFile("./testdata/2020-12/meta/applicator.json")
		testingutil.Expect(t, err, testingutil.Be[error](nil))

		p := &Payload{}
		err = p.UnmarshalJSON(data)
		testingutil.Expect(t, err, testingutil.Be[error](nil))

		p.Schema.PrintTo(os.Stdout)
	})
}

func TestSchema(t *testing.T) {
	t.Run("ref", func(t *testing.T) {
		testingutil.Expect(t, testingutil.MustJSONRaw(Any()), testingutil.Equal(`{"x-go-type":"any"}`))
	})

	t.Run("any", func(t *testing.T) {
		testingutil.Expect(t, testingutil.MustJSONRaw(Any()), testingutil.Equal(`{"x-go-type":"any"}`))
	})

	t.Run("string", func(t *testing.T) {
		testingutil.Expect(t, testingutil.MustJSONRaw(
			String(),
		), testingutil.
			Equal(`{"type":"string"}`),
		)
	})

	t.Run("bytes", func(t *testing.T) {
		testingutil.Expect(t, testingutil.MustJSONRaw(Bytes()), testingutil.Equal(`{"type":"string","format":"bytes"}`))
	})

	t.Run("binary", func(t *testing.T) {
		testingutil.Expect(t, testingutil.MustJSONRaw(Binary()), testingutil.Equal(`{"type":"string","format":"binary"}`))
	})

	t.Run("boolean", func(t *testing.T) {
		testingutil.Expect(t, testingutil.MustJSONRaw(Boolean()), testingutil.Equal(`{"type":"boolean"}`))
	})

	t.Run("array", func(t *testing.T) {
		testingutil.Expect(t, testingutil.MustJSONRaw(ArrayOf(String())), testingutil.Equal(`{"type":"array","items":{"type":"string"}}`))
	})

	t.Run("object", func(t *testing.T) {
		testingutil.Expect(t, testingutil.MustJSONRaw(
			ObjectOf(
				map[string]Schema{
					"key1": String(),
					"key2": String(),
				},
				"key1",
			),
		), testingutil.
			Equal(`{"type":"object","properties":{"key1":{"type":"string"},"key2":{"type":"string"}},"required":["key1"]}`),
		)
	})

	t.Run("object with additional", func(t *testing.T) {
		testingutil.Expect(t, testingutil.MustJSONRaw(
			MapOf(String()),
		), testingutil.
			Equal(`{"type":"object","additionalProperties":{"type":"string"}}`),
		)
	})

	t.Run("object with additionalProperties and propNames", func(t *testing.T) {
		testingutil.Expect(t, testingutil.MustJSONRaw(
			RecordOf(String(), String()),
		), testingutil.
			Equal(`{"type":"object","propertyNames":{"type":"string"},"additionalProperties":{"type":"string"}}`),
		)
	})

	t.Run("oneOf", func(t *testing.T) {
		testingutil.Expect(t, testingutil.MustJSONRaw(
			OneOf(String(), Boolean()),
		), testingutil.
			Equal(`{"oneOf":[{"type":"string"},{"type":"boolean"}]}`),
		)
	})
}

func TestCommon(t *testing.T) {
	data, err := json.Marshal(Metadata{Ext: ExtOf(map[string]any{
		"x-v": "string",
	})})

	testingutil.Expect(t, err, testingutil.Be[error](nil))
	testingutil.Expect(t, string(data), testingutil.Equal(`{"x-v":"string"}`))
}
