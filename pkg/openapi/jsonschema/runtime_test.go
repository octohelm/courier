package jsonschema

import (
	"bytes"
	"testing"

	"github.com/go-json-experiment/json"
	. "github.com/octohelm/x/testing/v2"
)

func TestAdditionalSchemaTypes(t *testing.T) {
	Then(t, "补充分支覆盖的 schema 类型行为正确",
		ExpectMust(func() error {
			buf := bytes.NewBuffer(nil)
			EnumType{Enum: []any{"a", 1}}.PrintTo(buf)
			if buf.String() != `"a" | 1` {
				return errSchemaCoverage("unexpected enum print: " + buf.String())
			}

			buf.Reset()
			AllOf(Any(), String(), Integer()).PrintTo(buf)
			if buf.String() != " & string & int" {
				return errSchemaCoverage("unexpected allOf print: " + buf.String())
			}

			buf.Reset()
			NullType{}.PrintTo(buf)
			if buf.String() != "null" {
				return errSchemaCoverage("unexpected null print")
			}

			buf.Reset()
			Long().(*NumberType).PrintTo(buf)
			if buf.String() != "int" {
				return errSchemaCoverage("unexpected long print")
			}

			buf.Reset()
			NumberType{Type: "number"}.PrintTo(buf)
			if buf.String() != "number" {
				return errSchemaCoverage("unexpected number print")
			}
			return nil
		}),
		ExpectMust(func() error {
			dynamic, _ := ParseURIReferenceString("#/$defs/User")
			ref := RefType{DynamicRef: dynamic}
			if ref.RefName() != "User" {
				return errSchemaCoverage("unexpected dynamic ref name")
			}

			buf := bytes.NewBuffer(nil)
			ref.PrintTo(buf)
			if buf.String() != "#User" {
				return errSchemaCoverage("unexpected dynamic ref print")
			}

			if (RefType{}).RefName() != "invalid" {
				return errSchemaCoverage("unexpected invalid ref name")
			}
			return nil
		}),
	)
}

func TestObjectPropsAndMetadata(t *testing.T) {
	Then(t, "对象属性与 metadata 行为正确",
		ExpectMust(func() error {
			var props Props
			if err := json.Unmarshal([]byte(`{"name":{"type":"string"},"age":{"type":"integer"}}`), &props); err != nil {
				return err
			}
			if props.Len() != 2 {
				return errSchemaCoverage("unexpected props len")
			}

			data, err := json.Marshal(props)
			if err != nil {
				return err
			}
			if !bytes.Contains(data, []byte(`"name"`)) || !bytes.Contains(data, []byte(`"age"`)) {
				return errSchemaCoverage("unexpected props marshal")
			}
			return nil
		}),
		ExpectMust(func() error {
			obj := ObjectOf(map[string]Schema{
				"name": String(),
			}, "name")
			obj.AdditionalProperties = Integer()
			obj.Defs = map[string]Schema{
				"Nested": &ObjectType{
					Type: "object",
					Core: Core{DynamicAnchor: "Nested"},
				},
			}
			meta := &Metadata{
				Title: "demo",
				Ext:   ExtOf(map[string]any{"x-demo": 1}),
			}
			cp := meta.DeepCopy()
			if cp == nil || cp.Title != "demo" {
				return errSchemaCoverage("unexpected deepcopy metadata")
			}

			buf := bytes.NewBuffer(nil)
			obj.PrintTo(buf, PrintWithDoc())
			if !bytes.Contains(buf.Bytes(), []byte(`#Nested`)) {
				return errSchemaCoverage("unexpected object print defs")
			}
			if !bytes.Contains(buf.Bytes(), []byte(`[X=string]: int`)) {
				return errSchemaCoverage("unexpected additional properties print")
			}
			return nil
		}),
	)
}

func errSchemaCoverage(msg string) error {
	return &schemaCoverageErr{msg: msg}
}

type schemaCoverageErr struct{ msg string }

func (e *schemaCoverageErr) Error() string { return e.msg }
