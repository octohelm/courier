package expression

import (
	testingx "github.com/octohelm/x/testing"
	"testing"
)

func TestParse(t *testing.T) {
	ex, _ := ParseString(`
select(
	when(
		pipe(get('x'), eq(1.1)),
		eq(1),
	),
	eq(2),
)
`)
	testingx.Expect(t, Stringify(ex), testingx.Be(`select(when(pipe(get("x"),eq(1.1)),eq(1)),eq(2))`))
}
