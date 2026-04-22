package courier

import (
	"context"
	"fmt"
	"testing"

	. "github.com/octohelm/x/testing/v2"
)

type DoSomeThing struct {
	Param int
}

func (req *DoSomeThing) SetDefaults() {
	if req != nil {
		if req.Param == 0 {
			req.Param = 1
		}
	}
}

func (DoSomeThing) Output(ctx context.Context) (any, error) {
	return nil, nil
}

func TestNewOperatorFactory(t *testing.T) {
	opInfo := NewOperatorFactory(&DoSomeThing{}, true)

	op := opInfo.New()

	Then(t, "新建 operator 时会应用默认值", Expect(op.(*DoSomeThing).Param, Equal(1)))
}

func Try(f func()) (err error) {
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("%v", e)
		}
	}()
	f()
	return err
}
