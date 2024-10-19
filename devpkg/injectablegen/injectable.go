package injectablegen

import (
	"go/types"
	"reflect"
	"strings"
	"sync"

	"github.com/octohelm/gengo/pkg/gengo"
)

func init() {
	gengo.Register(&injectableGen{})
}

type injectableGen struct {
	publicProviderInterface *types.Interface
	publicInitInterface     *types.Interface
	once                    sync.Once
}

func (*injectableGen) Name() string {
	return "injectable"
}

func (g *injectableGen) init(c gengo.Context) {
	{
		sig := c.Package("context").Function("Cause").Signature()

		g.publicProviderInterface = types.NewInterfaceType([]*types.Func{
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
		c.Render(gengo.Snippet{
			gengo.T: `
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
`,
			"Type":             gengo.ID(t.Obj()),
			"contextContext":   gengo.ID("context.Context"),
			"contextWithValue": gengo.ID("context.WithValue"),
		})
	case *types.Struct:
		provideFields := func(sw gengo.SnippetWriter) {
			if forAlias {
				return
			}

			for i := 0; i < x.NumFields(); i++ {
				f := x.Field(i)
				structTag := reflect.StructTag(x.Tag(i))

				injectTag, exists := structTag.Lookup("provide")
				if exists && injectTag != "-" {
					typ := f.Type()
					for {
						x, ok := typ.(*types.Pointer)
						if !ok {
							break
						}
						typ = x.Elem()
					}

					sw.Render(gengo.Snippet{
						gengo.T: `
ctx = @FieldType'InjectContext(ctx, p.@Field)
`,
						"Field":     gengo.ID(f.Name()),
						"FieldType": gengo.ID(f.Type()),
					})
				}
			}
		}

		if impl != "" {
			if !forAlias {
				c.Render(gengo.Snippet{
					gengo.T: `
func (p *@Type) InjectContext(ctx @contextContext) (@contextContext) {
   @provideFields		
   return @injectContext(ctx, p)
}

`,
					"Type":           gengo.ID(t.Obj()),
					"injectContext":  gengo.ID(impl + "InjectContext"),
					"contextContext": gengo.ID("context.Context"),
					"provideFields":  provideFields,
				})
			}

			return nil
		}

		c.Render(gengo.Snippet{
			gengo.T: `
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
`,
			"Type":             gengo.ID(t.Obj()),
			"contextContext":   gengo.ID("context.Context"),
			"contextWithValue": gengo.ID("context.WithValue"),
		})

		if !forAlias {
			c.Render(gengo.Snippet{
				gengo.T: `
func (p *@Type) InjectContext(ctx @contextContext) (@contextContext) {
   @provideFields
   return @Type'InjectContext(ctx, p)
}
`,
				"Type":             gengo.ID(t.Obj()),
				"contextContext":   gengo.ID("context.Context"),
				"contextWithValue": gengo.ID("context.WithValue"),
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

	c.Render(gengo.Snippet{
		gengo.T: `
func(v *@Type) Init(ctx @contextContext) error {
   @injectableFields		

   return nil
}

`,
		"Type":           gengo.ID(t.Obj()),
		"contextContext": gengo.ID("context.Context"),
		"injectableFields": func(sw gengo.SnippetWriter) {
			for i := 0; i < structType.NumFields(); i++ {
				f := structType.Field(i)
				structTag := reflect.StructTag(structType.Tag(i))

				injectTag, exists := structTag.Lookup("inject")
				if exists && injectTag != "-" {
					typ := f.Type()

					for {
						x, ok := typ.(*types.Pointer)
						if !ok {
							break
						}
						typ = x.Elem()
					}

					sw.Render(gengo.Snippet{
						gengo.T: `
if value, ok := @FieldType'FromContext(ctx); ok {
	v.@Field = value
} @elseOr
`,
						"Field":     gengo.ID(f.Name()),
						"FieldType": gengo.ID(typ),
						"elseOr": func(sw gengo.SnippetWriter) {
							if !strings.Contains(injectTag, ",optional") {
								sw.Render(gengo.Snippet{
									gengo.T: `else {
return @errorsErrorf("missing provider %T.@Field", v)
}
`,
									"Field":        gengo.ID(f.Name()),
									"errorsErrorf": gengo.ID("fmt.Errorf"),
								})
							}
						},
					})
				}
			}

			if g.hasBeforeInit(c, t.Obj().Pkg(), types.NewPointer(t.Obj().Type())) {
				sw.Render(gengo.Snippet{
					gengo.T: `
if err := v.beforeInit(ctx); err != nil {
	return err
}
`,
				})
			}

			for i := 0; i < structType.NumFields(); i++ {
				f := structType.Field(i)
				structTag := reflect.StructTag(structType.Tag(i))

				if _, exists := structTag.Lookup("inject"); exists {
					continue
				}

				if g.hasPublicInit(c, f.Type()) {
					sw.Render(gengo.Snippet{
						gengo.T: `
if err := v.@Field.Init(ctx); err != nil {
	return err
}
`,
						"Field": gengo.ID(f.Name()),
					})
				}
			}

			if g.hasAfterInit(c, t.Obj().Pkg(), types.NewPointer(t.Obj().Type())) {
				sw.Render(gengo.Snippet{
					gengo.T: `
if err := v.afterInit(ctx); err != nil {
	return err
}
`,
				})
			}
		},
	})
	return nil
}

func (g *injectableGen) isInjectable(c gengo.Context, t types.Type) bool {
	switch x := t.(type) {
	case *types.Pointer:
		return g.isInjectable(c, x.Elem())
	case *types.Named:
		tags, _ := c.Doc(x.Obj())
		if _, ok := tags["gengo:injectable"]; ok {
			return true
		}
		return types.Implements(x, g.publicProviderInterface) || types.Implements(types.NewPointer(x), g.publicProviderInterface)
	}

	return false
}

func (g *injectableGen) hasPublicInit(c gengo.Context, t types.Type) bool {
	switch x := t.(type) {
	case *types.Pointer:
		return g.hasPublicInit(c, x.Elem())
	case *types.Named:
		tags, _ := c.Doc(x.Obj())
		if _, ok := tags["gengo:injectable:provider"]; ok {
			_, ok := x.Obj().Type().(*types.Struct)
			return ok
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
