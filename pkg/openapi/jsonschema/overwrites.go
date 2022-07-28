package jsonschema

const (
	XEnumLabels   = `x-enum-labels`
	XGoVendorType = `x-go-vendor-type`
	XGoStarLevel  = `x-go-star-level`
	XGoFieldName  = `x-go-field-name`

	XTagValidate = `x-tag-validate`
)

type OpenAPISchemaTypeGetter interface {
	OpenAPISchemaType() []string
}

type OpenAPISchemaFormatGetter interface {
	OpenAPISchemaFormat() string
}
