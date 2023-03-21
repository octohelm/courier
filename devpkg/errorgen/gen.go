package errorgen

import (
	"go/types"

	"github.com/octohelm/gengo/pkg/gengo"
)

func init() {
	gengo.Register(&errorGen{})
}

type errorGen struct {
	errs map[types.Type]*ErrorType
}

type ErrorType struct {
	Constants []*types.Const
	Comments  map[*types.Const][]string
	ErrTalks  map[*types.Const]bool
}

func (*errorGen) Name() string {
	return "error"
}

func (g *errorGen) GenerateType(c gengo.Context, named *types.Named) error {
	if named == nil {
		return nil
	}

	p := c.Package(named.Obj().Pkg().Path())
	constants := p.Constants()
	for k := range p.Constants() {
		constv := constants[k]
		if named.Obj().Type().String() != constv.Type().String() {
			continue
		}

		if g.errs == nil {
			g.errs = make(map[types.Type]*ErrorType)
		}

		if g.errs[constv.Type()] == nil {
			g.errs[constv.Type()] = &ErrorType{
				Constants: make([]*types.Const, 0),
				Comments:  make(map[*types.Const][]string),
				ErrTalks:  make(map[*types.Const]bool),
			}
		}

		doc, comments := p.Doc(constv.Pos())
		errTalk := false
		if doc["errTalk"] != nil {
			comments = doc["errTalk"]
			errTalk = true
		}

		g.errs[constv.Type()].Constants = append(g.errs[constv.Type()].Constants, constv)
		g.errs[constv.Type()].Comments[constv] = comments
		g.errs[constv.Type()].ErrTalks[constv] = errTalk
	}

	g.genError(c, named)
	return nil
}

func (g *errorGen) genError(c gengo.Context, named *types.Named) {
	if g.errs[named.Obj().Type()] == nil {
		return
	}

	errs := g.errs[named.Obj().Type()]
	c.Render(gengo.Snippet{
		gengo.T: `
var _ interface {
	@StatusError
} = (*@Type)(nil)

func (v @Type) StatusErr() *@StatusErr {
	return &@StatusErr{
		Key:            v.Key(),
		Code:           v.Code(),
		Msg:            v.Msg(),
		CanBeTalkError: v.CanBeTalkError(),
	}
}

func (v  @Type) Unwrap() error {
	return v.StatusErr()
}

func (v  @Type) Error() string {
	return v.StatusErr().Error()
}

func (v  @Type) StatusCode() int {
	return @StatusCodeFromCode(int(v))
}

func (v @Type) Code() int {
	return int(v)
}

func (v @Type) Key() string {
	switch v {
		@constToKey
	}
	return "UNKNOWN"
}

func (v @Type) Msg() string {
	switch v {
		@constToMsg
	}
	return "-"
}

func (v @Type) CanBeTalkError() bool {
	switch v {
		@constToTaskError
	}
	return false
}
`,
		"Type":               gengo.ID(named.Obj().Type()),
		"StatusErr":          gengo.ID("github.com/octohelm/courier/pkg/statuserror.StatusErr"),
		"StatusError":        gengo.ID("github.com/octohelm/courier/pkg/statuserror.StatusError"),
		"StatusCodeFromCode": gengo.ID("github.com/octohelm/courier/pkg/statuserror.StatusCodeFromCode"),
		"constToKey": gengo.MapSnippet(errs.Constants, func(constkv *types.Const) gengo.Snippet {
			return gengo.Snippet{
				gengo.T: `
		case @ConstName:
			return @strValue
		`,
				"strValue":  constkv.Name(),
				"ConstName": gengo.ID(constkv.Id()),
			}
		}),
		"constToMsg": gengo.MapSnippet(errs.Constants, func(constkv *types.Const) gengo.Snippet {
			doc := ""
			if len(errs.Comments[constkv]) > 0 {
				doc = errs.Comments[constkv][0]
			}
			return gengo.Snippet{
				gengo.T: `
		case @ConstName:
			return @strValue
		`,
				"strValue":  doc,
				"ConstName": gengo.ID(constkv.Id()),
			}
		}),
		"constToTaskError": gengo.MapSnippet(errs.Constants, func(constkv *types.Const) gengo.Snippet {
			errTalk := "false"
			if errs.ErrTalks != nil && errs.ErrTalks[constkv] {
				errTalk = "true"
			}

			return gengo.Snippet{
				gengo.T: `
		case @ConstName:
			return @strValue
		`,
				"strValue":  gengo.ID(errTalk),
				"ConstName": gengo.ID(constkv.Id()),
			}
		}),
	})
}
