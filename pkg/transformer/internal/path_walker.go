package internal

import (
	"fmt"
	"strings"
)

func NewPathWalker() PathWalker {
	return &pathWalker{path: []any{}}
}

type PathWalker interface {
	Enter(i any)
	Exit()
	Paths() []any
	String() string
}

type pathWalker struct {
	path []any
}

func (pw *pathWalker) Enter(i any) {
	pw.path = append(pw.path, i)
}

func (pw *pathWalker) Exit() {
	pw.path = pw.path[:len(pw.path)-1]
}

func (pw *pathWalker) Paths() []any {
	return pw.path
}

func (pw *pathWalker) String() string {
	b := &strings.Builder{}
	for i := 0; i < len(pw.path); i++ {
		switch x := pw.path[i].(type) {
		case string:
			if b.Len() != 0 {
				b.WriteByte('.')
			}
			b.WriteString(x)
		case int:
			b.WriteString(fmt.Sprintf("[%d]", x))
		}
	}
	return b.String()
}
