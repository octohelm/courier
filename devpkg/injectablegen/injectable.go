package injectablegen

import (
	"go/types"
	"reflect"
	"strings"
	"sync"

	"github.com/octohelm/gengo/pkg/gengo/snippet"

	"github.com/octohelm/gengo/pkg/gengo"
)

func init() {
	gengo.Register(&injectableGen{})
}

type injectableGen struct {
	publicInjectContextInterface *types.Interface
	publicInitInterface          *types.Interface
	once                         sync.Once
}

func (*injectableGen) Name() string {
	return "injectable"
}

func (g *injectableGen) init(c gengo.Context) {
	{
		sig := c.Package("context").Function("Cause").Signature()

		g.publicInjectContextInterface = types.NewInterfaceType([]*types.Func{
			types.NewFunc(0, c.Package("context").Pkg(), "InjectContext",
				types.NewSignatureType(nil, nil, nil,
					types.NewTuple(sig.Params().At(0)),
					types.NewTuple(sig.Params().At(0)),
					false,
				),
			),
		}, nil)
	}

	{
		g.publicInitInterface = types.NewInterfaceType([]*types.Func{
			types.NewFunc(0, c.Package("context").Pkg(), "Init", c.Package("context").Function("Cause").Signature()),
		}, nil)
	}
}

func (g *injectableGen) GenerateType(c gengo.Context, t *types.Named) error {
	tags, _ := c.Doc(t.Obj())

	g.once.Do(func() {
		g.init(c)
	})

	values, ok := tags["gengo:injectable:provider"]
	if ok {
		if len(values) > 0 {
			if err := g.genAsProvider(c, t, values[0], false); err != nil {
				return err
			}
		} else {
			if err := g.genAsProvider(c, t, "", false); err != nil {
				return err
			}
		}
	}

	if err := g.genAsInjectable(c, t); err != nil {
		return err
	}

	return nil
}

func (g *injectableGen) GenerateAliasType(c gengo.Context, t *types.Alias) error {
	tags, _ := c.Doc(t.Obj())

	g.once.Do(func() {
		g.init(c)
	})

	values, ok := tags["gengo:injectable:provider"]
	if ok {
		if len(values) > 0 {
			if err := g.genAsProvider(c, t, values[0], true); err != nil {
				return err
			}
		} else {
			if err := g.genAsProvider(c, t, "", true); err != nil {
				return err
			}
		}
	}
	return nil
}

func (g *injectableGen) genAsProvider(c gengo.Context, t interface {
	Obj() *types.TypeName
	Underlying() types.Type
}, impl string, forAlias bool) error {
	switch x := t.Underlying().(type) {
	case *types.Interface:
		c.RenderT(`
type context@Type struct{}

func @Type'FromContext(ctx @contextContext) (@Type, bool) {
  if v, ok := ctx.Value(context@Type{}).(@Type); ok {
      return v, true
   }
   return nil, false
}

func @Type'InjectContext(ctx @contextContext, tpe @Type) (@contextContext) {
   return @contextWithValue(ctx, context@Type{}, tpe)
}
`, snippet.Args{
			"Type":             snippet.ID(t.Obj()),
			"contextContext":   snippet.ID("context.Context"),
			"contextWithValue": snippet.ID("context.WithValue"),
		})
	case *types.Struct:
		hasProvideFields := func() bool {
			for i := 0; i < x.NumFields(); i++ {
				structTag := reflect.StructTag(x.Tag(i))

				injectTag, exists := structTag.Lookup("provide")
				if exists && injectTag != "-" {
					return true
				}
			}

			return false
		}()

		provideFields := snippet.Snippets(func(yield func(snippet.Snippet) bool) {
			if forAlias {
				return
			}

			for i := 0; i < x.NumFields(); i++ {
				f := x.Field(i)
				structTag := reflect.StructTag(x.Tag(i))

				_, injectExists := structTag.Lookup("inject")

				injectTag, provideExists := structTag.Lookup("provide")
				if provideExists && injectTag != "-" {
					typ := f.Type()
					for {
						x, ok := typ.(*types.Pointer)
						if !ok {
							break
						}
						typ = x.Elem()
					}

					optional := strings.Contains(injectTag, ",opt")

					if optional {
						if !yield(snippet.T(`
if p.@Field != nil {
	ctx = @FieldType'InjectContext(ctx, p.@Field)
}
`, snippet.Args{
							"Field":     snippet.ID(f.Name()),
							"FieldType": snippet.ID(typ),
						})) {
							return
						}
					} else {
						if !yield(snippet.T(`
ctx = @FieldType'InjectContext(ctx, p.@Field)
`, snippet.Args{
							"Field":     snippet.ID(f.Name()),
							"FieldType": snippet.ID(typ),
						})) {
							return
						}
					}
				}

				if !injectExists && !provideExists {
					if g.hasPublicInjectContext(c, f.Type()) {
						if !yield(snippet.T(`
ctx = p.@Field.InjectContext(ctx)
`, snippet.Args{
							"Field": snippet.ID(f.Name()),
						})) {
							return
						}
						continue
					}
				}
			}
		})

		if impl != "" {
			if !forAlias {
				c.RenderT(`
func (p *@Type) InjectContext(ctx @contextContext) (@contextContext) {
   @provideFields		
   return @injectContext(ctx, p)
}

`, snippet.Args{
					"Type":           snippet.ID(t.Obj()),
					"injectContext":  snippet.ID(impl + "InjectContext"),
					"contextContext": snippet.ID("context.Context"),
					"provideFields":  provideFields,
				})
			}

			return nil
		}

		if !hasProvideFields {
			c.RenderT(`
type context@Type struct{}

func @Type'FromContext(ctx @contextContext) (*@Type, bool) {
  if v, ok := ctx.Value(context@Type{}).(*@Type); ok {
      return v, true
   }
   return nil, false
}

func @Type'InjectContext(ctx @contextContext, tpe *@Type) (@contextContext) {
   return @contextWithValue(ctx, context@Type{}, tpe)
}

`, snippet.Args{
				"Type":             snippet.ID(t.Obj()),
				"contextContext":   snippet.ID("context.Context"),
				"contextWithValue": snippet.ID("context.WithValue"),
			})
		}

		if !forAlias {
			if hasProvideFields {
				c.RenderT(`
func (p *@Type) InjectContext(ctx @contextContext) (@contextContext) {
   @provideFields
   return ctx
}
`, snippet.Args{
					"Type":             snippet.ID(t.Obj()),
					"contextContext":   snippet.ID("context.Context"),
					"contextWithValue": snippet.ID("context.WithValue"),
					"provideFields":    provideFields,
				})

				return nil
			}

			c.RenderT(`
func (p *@Type) InjectContext(ctx @contextContext) (@contextContext) {
   @provideFields
   return @Type'InjectContext(ctx, p)
}
`, snippet.Args{
				"Type":             snippet.ID(t.Obj()),
				"contextContext":   snippet.ID("context.Context"),
				"contextWithValue": snippet.ID("context.WithValue"),
				"provideFields":    provideFields,
			})
		}
	}

	return nil
}

func (g *injectableGen) genAsInjectable(c gengo.Context, t *types.Named) error {
	structType, ok := t.Obj().Type().Underlying().(*types.Struct)
	if !ok {
		return nil
	}

	c.RenderT(`
func(v *@Type) Init(ctx @contextContext) error {
   @injectableFields		

   return nil
}

`, snippet.Args{
		"Type":           snippet.ID(t.Obj()),
		"contextContext": snippet.ID("context.Context"),
		"injectableFields": snippet.Snippets(func(yield func(snippet.Snippet) bool) {
			for i := 0; i < structType.NumFields(); i++ {
				f := structType.Field(i)
				structTag := reflect.StructTag(structType.Tag(i))

				injectTag, injectExists := structTag.Lookup("inject")
				if injectExists && injectTag != "-" {
					typ := f.Type()

					for {
						x, ok := typ.(*types.Pointer)
						if !ok {
							break
						}
						typ = x.Elem()
					}

					if !yield(snippet.T(`
if value, ok := @FieldType'FromContext(ctx); ok {
	v.@Field = value
} @elseOr
`, snippet.Args{
						"Field":     snippet.ID(f.Name()),
						"FieldType": snippet.ID(typ),
						"elseOr": snippet.Snippets(func(yield func(snippet.Snippet) bool) {
							if !strings.Contains(injectTag, ",opt") {
								if !yield(snippet.T(`else {
return @errorsErrorf("missing provider %T.@Field", v)
}
`, snippet.Args{
									"Field":        snippet.ID(f.Name()),
									"errorsErrorf": snippet.ID("fmt.Errorf"),
								})) {
									return
								}
							}
						}),
					})) {
						return
					}
				}
			}

			if g.hasBeforeInit(c, t.Obj().Pkg(), types.NewPointer(t.Obj().Type())) {
				if !yield(snippet.Block(`
if err := v.beforeInit(ctx); err != nil {
	return err
}
`)) {
					return
				}
			}

			for i := 0; i < structType.NumFields(); i++ {
				f := structType.Field(i)
				structTag := reflect.StructTag(structType.Tag(i))

				_, injectExists := structTag.Lookup("inject")
				_, provideExists := structTag.Lookup("provide")

				if injectExists || provideExists {
					continue
				}

				if g.hasPublicInit(c, f.Type()) {
					if !yield(snippet.T(`
if err := v.@Field.Init(ctx); err != nil {
	return err
}
`, snippet.Args{
						"Field": snippet.ID(f.Name()),
					})) {
						return
					}
				}
			}

			if g.hasAfterInit(c, t.Obj().Pkg(), types.NewPointer(t.Obj().Type())) {
				if !yield(snippet.Block(`
if err := v.afterInit(ctx); err != nil {
	return err
}
`)) {
					return
				}
			}
		}),
	})
	return nil
}

func (g *injectableGen) hasPublicInjectContext(c gengo.Context, t types.Type) bool {
	switch x := t.(type) {
	case *types.Pointer:
		return g.hasPublicInjectContext(c, x.Elem())
	case *types.Named:
		_, isStruct := x.Underlying().(*types.Struct)
		if !isStruct {
			return false
		}
		tags, _ := c.Doc(x.Obj())
		if _, ok := tags["gengo:injectable:provider"]; ok {
			return ok
		}
		return types.Implements(x, g.publicInjectContextInterface) || types.Implements(types.NewPointer(x), g.publicInjectContextInterface)
	}

	return false
}

func (g *injectableGen) hasPublicInit(c gengo.Context, t types.Type) bool {
	switch x := t.(type) {
	case *types.Pointer:
		return g.hasPublicInit(c, x.Elem())
	case *types.Named:
		_, isStruct := x.Underlying().(*types.Struct)
		if !isStruct {
			return false
		}
		tags, _ := c.Doc(x.Obj())
		_, injectable := tags["gengo:injectable"]
		_, injectableProvider := tags["gengo:injectable:provider"]
		if injectable || injectableProvider {
			return true
		}
		return types.Implements(x, g.publicInitInterface) || types.Implements(types.NewPointer(x), g.publicInitInterface)
	}

	return false
}

func (g *injectableGen) hasBeforeInit(c gengo.Context, pkg *types.Package, t types.Type) bool {
	switch x := t.(type) {
	case *types.Pointer:
		initInterface := types.NewInterfaceType([]*types.Func{
			types.NewFunc(0, pkg, "beforeInit", c.Package("context").Function("Cause").Signature()),
		}, nil)

		return types.Implements(x, initInterface)
	}
	return false
}

func (g *injectableGen) hasAfterInit(c gengo.Context, pkg *types.Package, t types.Type) bool {
	switch x := t.(type) {
	case *types.Pointer:
		initInterface := types.NewInterfaceType([]*types.Func{
			types.NewFunc(0, pkg, "afterInit", c.Package("context").Function("Cause").Signature()),
		}, nil)

		return types.Implements(x, initInterface)
	}
	return false
}
