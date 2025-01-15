package defaultergen

import (
	"go/types"

	"github.com/octohelm/gengo/pkg/gengo"
	"github.com/octohelm/gengo/pkg/gengo/snippet"
)

func init() {
	gengo.Register(&unitstr{})
}

type unitstr struct{}

func (*unitstr) Name() string {
	return "uintstr"
}

func (g *unitstr) GenerateType(c gengo.Context, t *types.Named) error {
	if b, ok := t.Obj().Type().Underlying().(*types.Basic); ok {
		switch b.Kind() {
		case types.Uint64, types.Uint32, types.Uint16, types.Uint8, types.Uint:
			c.RenderT(`
func(id *@Type) UnmarshalText(text []byte) error {
	str := string(text)
	if len(str) == 0 {
		return nil
	}
	v, err := @strconvParseUint(str, 10, 64)
	if err != nil {
		return err
	}
	*id = @Type(v)
	return nil
}

func (id @Type) MarshalText() (text []byte, err error) {
	if id == 0 {
		return nil, nil
	}
	return []byte(@strconvFormatUint(uint64(id), 10)), nil
}


func (id @Type) String() string {
	return @strconvFormatUint(uint64(id), 10)
}

`, snippet.Args{
				"Type":              snippet.ID(t.Obj()),
				"strconvParseUint":  snippet.PkgExpose("strconv", "ParseUint"),
				"strconvFormatUint": snippet.PkgExpose("strconv", "FormatUint"),
			})
		default:

		}
	}

	return nil
}
