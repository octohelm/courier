package courier

import (
	"context"
	"fmt"
	"testing"

	testingx "github.com/octohelm/x/testing"
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

	testingx.Expect(t, op.(*DoSomeThing).Param, testingx.Equal(1))
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
