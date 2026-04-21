package jsonschema

import (
	"bytes"
	"testing"

	"github.com/go-json-experiment/json"
	. "github.com/octohelm/x/testing/v2"
)

func TestSchemaHelpers(t0 *testing.T) {
	Then(t0, "schema 公开 helper 行为正确",
		Expect(IsType[*StringType](String()), Equal(true)),
		Expect(IsType[*StringType](Integer()), Equal(false)),
		ExpectMust(func() error {
			s := ObjectOf(map[string]Schema{
				"name": String(),
				"id":   Integer(),
			}, "id")
			if s.Type != "object" || len(s.Required) != 1 || s.Required[0] != "id" {
				return errSchema("unexpected object type")
			}
			return nil
		}),
		ExpectMust(func() error {
			record := RecordOf(String(), Integer())
			m := MapOf(String())
			if record.PropertyNames == nil || record.AdditionalProperties == nil || m.AdditionalProperties == nil {
				return errSchema("unexpected record/map schema")
			}
			return nil
		}),
		ExpectMust(func() error {
			data, err := json.Marshal(Payload{Schema: String()})
			if err != nil {
				return err
			}
			var payload Payload
			if err := json.Unmarshal(data, &payload); err != nil {
				return err
			}
			if !IsType[*StringType](payload.Schema) {
				return errSchema("unexpected payload schema type")
			}
			return nil
		}),
		ExpectMust(func() error {
			var payload Payload
			if err := Unmarshal([]byte(`{"type":"integer"}`), &payload); err != nil {
				return err
			}
			if !IsType[*NumberType](payload.Schema) {
				return errSchema("unexpected unmarshal schema type")
			}
			return nil
		}),
	)
}

func TestMetadataAndStrfmtHelpers(t0 *testing.T) {
	Then(t0, "metadata 与 strfmt helper 行为正确",
		ExpectMust(func() error {
			meta := &Metadata{}
			if meta.GetMetadata() != meta {
				return errSchema("unexpected metadata getter")
			}
			ext := ExtOf(map[string]any{"x-a": 1})
			if v, ok := ext.GetExtension("x-a"); !ok || v.(int) != 1 {
				return errSchema("unexpected ext value")
			}
			ext.AddExtension("x-b", 2)
			ext.AddExtension("x-c", nil)
			if _, ok := ext.GetExtension("x-b"); !ok {
				return errSchema("missing extension")
			}
			copy := ext.DeepCopy()
			if copy == nil || copy.Extensions["x-a"].(int) != 1 {
				return errSchema("unexpected deepcopy")
			}
			merged := ext.Merge(ExtOf(map[string]any{"x-z": 9}))
			if _, ok := merged.GetExtension("x-a"); !ok {
				return errSchema("missing merged source extension")
			}
			if v, ok := merged.GetExtension("x-z"); ok && v != nil {
				return errSchema("merge behavior changed unexpectedly")
			}
			return nil
		}),
		ExpectMust(func() error {
			var uri URIString
			if err := uri.UnmarshalText([]byte("https://example.com/path?q=1")); err != nil {
				return err
			}
			raw, err := uri.MarshalText()
			if err != nil || string(raw) != "https://example.com/path?q=1" {
				return errSchema("unexpected uri marshal")
			}

			ref, err := ParseURIReferenceString("/components/schemas/Pet#/$defs/Pet")
			if err != nil {
				return err
			}
			if ref.RefName() != "Pet" {
				return errSchema("unexpected ref name")
			}
			raw, err = ref.MarshalText()
			if err != nil || string(raw) != "/components/schemas/Pet#/$defs/Pet" {
				return errSchema("unexpected ref marshal")
			}

			ref2 := URIReferenceString{Scheme: "https", Host: "example.com", Path: "/a"}
			raw, err = ref2.MarshalText()
			if err != nil || string(raw) != "https://example.com/a" {
				return errSchema("unexpected absolute ref marshal")
			}

			var anchor AnchorString
			if err := anchor.UmarshalText([]byte("valid_anchor")); err == nil {
				return errSchema("expected current anchor validation behavior")
			}
			if !IsType[*StringType](anchor.OpenAPISchema()) {
				return errSchema("unexpected anchor schema")
			}
			return nil
		}),
	)
}

func TestPrinterAndObjectHelpers(t0 *testing.T) {
	Then(t0, "printer、ref 与 object helper 行为正确",
		ExpectMust(func() error {
			opt := &printOption{}
			opt.Build(PrintWithDoc())
			if !opt.WithDoc {
				return errSchema("unexpected print option")
			}

			buf := bytes.NewBuffer(nil)
			Print(buf, func(p Printer) {
				p.PrintDoc("line1\nline2")
				p.Print("{")
				p.Return()
				end := p.Indent()
				p.Print("x")
				end()
				p.Return()
				p.Print("}")
			})
			if !bytes.Contains(buf.Bytes(), []byte("// line1")) || !bytes.Contains(buf.Bytes(), []byte("// line2")) {
				return errSchema("unexpected printer doc output")
			}
			return nil
		}),
		ExpectMust(func() error {
			refName, _ := ParseURIReferenceString("#/$defs/Pet")
			ref := RefType{Ref: refName}
			buf := bytes.NewBuffer(nil)
			ref.PrintTo(buf)
			if ref.RefName() != "Pet" || buf.String() != "#Pet" {
				return errSchema("unexpected ref output")
			}
			buf.Reset()
			RefType{}.PrintTo(buf)
			if buf.String() != "_|_" {
				return errSchema("unexpected invalid ref output")
			}
			return nil
		}),
		ExpectMust(func() error {
			props := Props{}
			if !props.IsZero() {
				return errSchema("expected zero props")
			}
			if !props.Set("name", String()) || props.Set("name", Integer()) {
				return errSchema("unexpected props set behavior")
			}
			if props.Len() != 1 {
				return errSchema("unexpected props len")
			}
			if s, ok := props.Get("name"); !ok || !IsType[*NumberType](s) {
				return errSchema("unexpected props get")
			}
			if !props.Delete("name") || props.Delete("name") {
				return errSchema("unexpected props delete")
			}
			return nil
		}),
		ExpectMust(func() error {
			obj := ObjectOf(map[string]Schema{"name": String()}, "name")
			obj.SetProperty("age", Integer(), false)
			if _, ok := obj.Properties.Get("age"); !ok || len(obj.Required) != 1 || obj.Required[0] != "name" {
				return errSchema("unexpected object set property result")
			}
			return nil
		}),
	)
}

func TestJSONSchemaRuntimeDoc(t0 *testing.T) {
	Then(t0, "generated runtime doc 可以访问",
		ExpectMust(func() error {
			if doc, ok := new(AnchorString).RuntimeDoc(); !ok || len(doc) == 0 {
				return errSchema("missing anchor runtime doc")
			}
			if _, ok := new(AnyType).RuntimeDoc("Schema"); !ok {
				return errSchema("missing any runtime field")
			}
			if _, ok := new(ArrayType).RuntimeDoc("MaxItems"); !ok {
				return errSchema("missing array runtime field")
			}
			if _, ok := new(BooleanType).RuntimeDoc("Type"); !ok {
				return errSchema("missing boolean runtime field")
			}
			if _, ok := new(Core).RuntimeDoc("Anchor"); !ok {
				return errSchema("missing core runtime field")
			}
			if _, ok := new(Discriminator).RuntimeDoc("Mapping"); !ok {
				return errSchema("missing discriminator runtime field")
			}
			if _, ok := new(EnumType).RuntimeDoc("Enum"); !ok {
				return errSchema("missing enum runtime field")
			}
			if _, ok := new(Ext).RuntimeDoc("Extensions"); !ok {
				return errSchema("missing ext runtime field")
			}
			if _, ok := new(IntersectionType).RuntimeDoc("AllOf"); !ok {
				return errSchema("missing intersection runtime field")
			}
			if _, ok := new(Metadata).RuntimeDoc("Title"); !ok {
				return errSchema("missing metadata runtime field")
			}
			if _, ok := new(NullType).RuntimeDoc("Type"); !ok {
				return errSchema("missing null runtime field")
			}
			if _, ok := new(NumberType).RuntimeDoc("MultipleOf"); !ok {
				return errSchema("missing number runtime field")
			}
			if _, ok := new(ObjectType).RuntimeDoc("Properties"); !ok {
				return errSchema("missing object runtime field")
			}
			if _, ok := new(RefType).RuntimeDoc("Ref"); !ok {
				return errSchema("missing ref runtime field")
			}
			if _, ok := new(StringType).RuntimeDoc("Pattern"); !ok {
				return errSchema("missing string runtime field")
			}
			if _, ok := new(UnionType).RuntimeDoc("OneOf"); !ok {
				return errSchema("missing union runtime field")
			}
			if _, ok := runtimeDoc(struct{}{}, "", "Type"); ok {
				return errSchema("unexpected runtimeDoc hit")
			}
			return nil
		}),
	)
}

func errSchema(msg string) error {
	return &schemaErr{msg: msg}
}

type schemaErr struct{ msg string }

func (e *schemaErr) Error() string { return e.msg }
