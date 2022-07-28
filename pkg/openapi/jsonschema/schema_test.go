package jsonschema

import (
	"regexp"
	"testing"

	"github.com/octohelm/courier/internal/testingutil"
	"github.com/octohelm/courier/pkg/ptr"
	"github.com/onsi/gomega"
)

func TestSchema(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		gomega.NewWithT(t).Expect(testingutil.MustJSONRaw(Schema{})).To(gomega.Equal(`{}`))
	})

	t.Run("integer", func(t *testing.T) {
		gomega.NewWithT(t).Expect(testingutil.MustJSONRaw(Integer())).To(gomega.Equal(`{"type":"integer","format":"int32"}`))
	})

	t.Run("long", func(t *testing.T) {
		gomega.NewWithT(t).Expect(testingutil.MustJSONRaw(Long())).To(gomega.Equal(`{"type":"integer","format":"int64"}`))
	})

	t.Run("float", func(t *testing.T) {
		gomega.NewWithT(t).Expect(testingutil.MustJSONRaw(Float())).To(gomega.Equal(`{"type":"number","format":"float"}`))
	})

	t.Run("double", func(t *testing.T) {
		gomega.NewWithT(t).Expect(testingutil.MustJSONRaw(Double())).To(gomega.Equal(`{"type":"number","format":"double"}`))
	})

	t.Run("string", func(t *testing.T) {
		gomega.NewWithT(t).Expect(testingutil.MustJSONRaw(
			String().WithTitle("title").WithDesc("desc"),
		)).To(
			gomega.Equal(`{"type":"string","title":"title","description":"desc"}`),
		)
	})

	t.Run("bytes", func(t *testing.T) {
		gomega.NewWithT(t).Expect(testingutil.MustJSONRaw(Bytes())).To(gomega.Equal(`{"type":"string","format":"bytes"}`))
	})

	t.Run("binary", func(t *testing.T) {
		gomega.NewWithT(t).Expect(testingutil.MustJSONRaw(Binary())).To(gomega.Equal(`{"type":"string","format":"binary"}`))
	})

	t.Run("boolean", func(t *testing.T) {
		gomega.NewWithT(t).Expect(testingutil.MustJSONRaw(Boolean())).To(gomega.Equal(`{"type":"boolean"}`))
	})

	t.Run("array", func(t *testing.T) {
		gomega.NewWithT(t).Expect(testingutil.MustJSONRaw(ItemsOf(String()))).To(gomega.Equal(`{"type":"array","items":{"type":"string"}}`))
	})

	t.Run("object", func(t *testing.T) {
		gomega.NewWithT(t).Expect(testingutil.MustJSONRaw(
			ObjectOf(
				Props{
					"key1": String(),
					"key2": String(),
				},
				"key1",
			),
		)).To(
			gomega.Equal(`{"type":"object","properties":{"key1":{"type":"string"},"key2":{"type":"string"}},"required":["key1"]}`),
		)
	})

	t.Run("object with additional", func(t *testing.T) {
		gomega.NewWithT(t).Expect(testingutil.MustJSONRaw(
			MapOf(String()),
		)).To(
			gomega.Equal(`{"type":"object","additionalProperties":{"type":"string"}}`),
		)
	})

	t.Run("object with additional and propNames", func(t *testing.T) {
		gomega.NewWithT(t).Expect(testingutil.MustJSONRaw(
			KeyValueOf(String(), String()),
		)).To(
			gomega.Equal(`{"type":"object","additionalProperties":{"type":"string"},"propertyNames":{"type":"string"}}`),
		)
	})

	t.Run("anyOf", func(t *testing.T) {
		gomega.NewWithT(t).Expect(testingutil.MustJSONRaw(
			AnyOf(String(), Boolean()),
		)).To(
			gomega.Equal(`{"anyOf":[{"type":"string"},{"type":"boolean"}]}`),
		)
	})

	t.Run("oneOf", func(t *testing.T) {
		gomega.NewWithT(t).Expect(testingutil.MustJSONRaw(
			OneOf(String(), Boolean()),
		)).To(
			gomega.Equal(`{"oneOf":[{"type":"string"},{"type":"boolean"}]}`),
		)
	})

	t.Run("allOf", func(t *testing.T) {
		gomega.NewWithT(t).Expect(testingutil.MustJSONRaw(
			AllOf(String(), Boolean()),
		)).To(
			gomega.Equal(`{"allOf":[{"type":"string"},{"type":"boolean"}]}`),
		)
	})

	t.Run("not", func(t *testing.T) {
		gomega.NewWithT(t).Expect(testingutil.MustJSONRaw(
			Not(String()),
		)).To(
			gomega.Equal(`{"not":{"type":"string"}}`),
		)
	})

	t.Run("Validation", func(t *testing.T) {
		validation := &SchemaValidation{
			MultipleOf:       ptr.Ptr(2.0),
			Maximum:          ptr.Ptr(10.0),
			ExclusiveMaximum: true,
			Minimum:          ptr.Ptr(1.0),
			ExclusiveMinimum: true,

			MaxLength: ptr.Ptr[uint64](10),
			MinLength: ptr.Ptr[uint64](0),
			Pattern:   regexp.MustCompile("/+d/").String(),

			MaxItems:    ptr.Ptr[uint64](10),
			MinItems:    ptr.Ptr[uint64](1),
			UniqueItems: true,

			MaxProperties: ptr.Ptr[uint64](10.0),
			MinProperties: ptr.Ptr[uint64](1.0),
			Required:      []string{"key"},

			Enum: []any{"1", "2", "3"},
		}

		t.Run("with string validation", func(t *testing.T) {
			gomega.NewWithT(t).Expect(testingutil.MustJSONRaw(
				String().WithValidation(validation),
			)).To(
				gomega.Equal(`{"type":"string","maxLength":10,"minLength":0,"pattern":"/+d/","enum":["1","2","3"]}`),
			)
		})

		t.Run("with integer validation", func(t *testing.T) {
			gomega.NewWithT(t).Expect(testingutil.MustJSONRaw(
				Integer().WithValidation(validation),
			)).To(
				gomega.Equal(`{"type":"integer","format":"int32","multipleOf":2,"maximum":10,"exclusiveMaximum":true,"minimum":1,"exclusiveMinimum":true,"enum":["1","2","3"]}`),
			)
		})

		t.Run(
			"with number validation", func(t *testing.T) {
				gomega.NewWithT(t).Expect(testingutil.MustJSONRaw(
					Float().WithValidation(validation),
				)).To(
					gomega.Equal(`{"type":"number","format":"float","multipleOf":2,"maximum":10,"exclusiveMaximum":true,"minimum":1,"exclusiveMinimum":true,"enum":["1","2","3"]}`),
				)
			})

		t.Run("with array validation", func(t *testing.T) {
			gomega.NewWithT(t).Expect(testingutil.MustJSONRaw(
				ItemsOf(String()).WithValidation(validation),
			)).To(
				gomega.Equal(`{"type":"array","items":{"type":"string"},"maxItems":10,"minItems":1,"uniqueItems":true,"enum":["1","2","3"]}`),
			)
		})

		t.Run("with object validation", func(t *testing.T) {
			gomega.NewWithT(t).Expect(testingutil.MustJSONRaw(
				ObjectOf(
					Props{
						"key1": String(),
						"key2": String(),
					},
					"key1",
				).WithValidation(validation),
			)).To(
				gomega.Equal(`{"type":"object","properties":{"key1":{"type":"string"},"key2":{"type":"string"}},"maxProperties":10,"minProperties":1,"required":["key"],"enum":["1","2","3"]}`),
			)
		})
	})

}
