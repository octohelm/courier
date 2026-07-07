package taggeduniongen

import (
	"go/types"
	"reflect"

	"github.com/octohelm/gengo/pkg/gengo"
	"github.com/octohelm/gengo/pkg/gengo/snippet"
)

func init() {
	gengo.Register(&taggedUnionGen{})
}

type taggedUnionGen struct{}

func (g *taggedUnionGen) Name() string {
	return "taggedunion"
}

func (g *taggedUnionGen) GenerateType(c gengo.Context, t *types.Named) error {
	structType, ok := t.Obj().Type().Underlying().(*types.Struct)
	if !ok {
		return gengo.ErrSkip
	}

	discriminator, fields, err := parseTaggedUnion(structType)
	if err != nil {
		return err
	}
	if discriminator == "" || len(fields) == 0 {
		return gengo.ErrSkip
	}

	opts := c.OptsOf(t.Obj(), "taggedunion")
	underlyingTypeName, _ := opts.Get("underlying")

	g.generateDiscriminator(c, t, discriminator)
	g.generateMapping(c, t, fields)
	g.generateSetUnderlying(c, t, fields)
	g.generateUnderlying(c, t, fields, underlyingTypeName)
	g.generateIsZero(c, t, fields)
	g.generateMarshalJSON(c, t)
	g.generateUnmarshalJSON(c, t)

	return nil
}

type mappingField struct {
	FieldName string
	TagValue  string
	FieldType types.Type
}

func parseTaggedUnion(structType *types.Struct) (string, []mappingField, error) {
	discriminator := ""
	var fields []mappingField

	for i := 0; i < structType.NumFields(); i++ {
		f := structType.Field(i)
		tag := reflect.StructTag(structType.Tag(i))

		if f.Embedded() {
			if named, ok := f.Type().(*types.Named); ok {
				if named.Obj().Pkg().Path() == "github.com/octohelm/courier/pkg/taggedunion" &&
					named.Obj().Name() == "TaggedUnion" {
					discriminator = tag.Get("discriminator")
				}
			}
			continue
		}

		mapping := tag.Get("mapping")
		if mapping == "" {
			continue
		}

		elemType := f.Type()
		if ptr, ok := elemType.(*types.Pointer); ok {
			elemType = ptr.Elem()
		}

		fields = append(fields, mappingField{
			FieldName: f.Name(),
			TagValue:  mapping,
			FieldType: elemType,
		})
	}

	return discriminator, fields, nil
}

func (g *taggedUnionGen) generateDiscriminator(c gengo.Context, t *types.Named, discriminator string) {
	c.RenderT(`
func (@Type) Discriminator() string {
	return @Discriminator
}

`, snippet.Args{
		"Type":          snippet.ID(t.Obj()),
		"Discriminator": snippet.Value(discriminator),
	})
}

func (g *taggedUnionGen) generateMapping(c gengo.Context, t *types.Named, fields []mappingField) {
	c.RenderT(`
func (@Type) Mapping() map[string]any {
	return map[string]any{
		@mappingEntries
	}
}

`, snippet.Args{
		"Type": snippet.ID(t.Obj()),
		"mappingEntries": snippet.Snippets(func(yield func(snippet.Snippet) bool) {
			for _, f := range fields {
				if !yield(snippet.T(`
		@Key: new(@FieldType),
`, snippet.Args{
					"Key":       snippet.Value(f.TagValue),
					"FieldType": snippet.ID(f.FieldType),
				})) {
					return
				}
			}
		}),
	})
}

func (g *taggedUnionGen) generateSetUnderlying(c gengo.Context, t *types.Named, fields []mappingField) {
	c.RenderT(`
func (u *@Type) SetUnderlying(v any) {
	switch x := v.(type) {
		@cases
	}
}

`, snippet.Args{
		"Type": snippet.ID(t.Obj()),
		"cases": snippet.Snippets(func(yield func(snippet.Snippet) bool) {
			for _, f := range fields {
				if !yield(snippet.T(`
	case *@FieldType:
		u.@Field = x
`, snippet.Args{
					"FieldType": snippet.ID(f.FieldType),
					"Field":     snippet.ID(f.FieldName),
				})) {
					return
				}
			}
		}),
	})
}

func (g *taggedUnionGen) generateUnderlying(c gengo.Context, t *types.Named, fields []mappingField, underlyingTypeName string) {
	returnType := "any"
	if underlyingTypeName != "" {
		returnType = underlyingTypeName
	}

	c.RenderT(`
func (u *@Type) Underlying() @ReturnType {
	@checks
	return nil
}

`, snippet.Args{
		"Type":       snippet.ID(t.Obj()),
		"ReturnType": snippet.ID(returnType),
		"checks": snippet.Snippets(func(yield func(snippet.Snippet) bool) {
			for _, f := range fields {
				if !yield(snippet.T(`
		if u.@Field != nil {
			return u.@Field
		}
`, snippet.Args{
					"Field": snippet.ID(f.FieldName),
				})) {
					return
				}
			}
		}),
	})
}

func (g *taggedUnionGen) generateIsZero(c gengo.Context, t *types.Named, fields []mappingField) {
	c.RenderT(`
func (u *@Type) IsZero() bool {
	return @nilChecks
}

`, snippet.Args{
		"Type": snippet.ID(t.Obj()),
		"nilChecks": snippet.Snippets(func(yield func(snippet.Snippet) bool) {
			last := len(fields) - 1
			for i, f := range fields {
				if !yield(snippet.T(`u.@Field == nil`, snippet.Args{
					"Field": snippet.ID(f.FieldName),
				})) {
					return
				}
				if i < last {
					if !yield(snippet.Block(" && ")) {
						return
					}
				}
			}
		}),
	})
}

func (g *taggedUnionGen) generateMarshalJSON(c gengo.Context, t *types.Named) {
	c.RenderT(`
func (u @Type) MarshalJSON() ([]byte, error) {
	if u.Underlying() == nil {
		return []byte("{}"), nil
	}
	return @validatorMarshal(u.Underlying())
}

`, snippet.Args{
		"Type":             snippet.ID(t.Obj()),
		"validatorMarshal": snippet.ID("github.com/octohelm/courier/pkg/validator.Marshal"),
	})
}

func (g *taggedUnionGen) generateUnmarshalJSON(c gengo.Context, t *types.Named) {
	c.RenderT(`
func (u *@Type) UnmarshalJSON(data []byte) error {
	mm := @Type{}
	if err := @unmarshalDecode(@jsontextNewDecoder(@bytesNewBuffer(data)), &mm); err != nil {
		return err
	}
	*u = mm
	return nil
}

`, snippet.Args{
		"Type":               snippet.ID(t.Obj()),
		"unmarshalDecode":    snippet.ID("github.com/octohelm/courier/pkg/validator/taggedunion.UnmarshalDecode"),
		"jsontextNewDecoder": snippet.ID("github.com/go-json-experiment/json/jsontext.NewDecoder"),
		"bytesNewBuffer":     snippet.ID("bytes.NewBuffer"),
	})
}
