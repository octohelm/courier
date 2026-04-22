package expression

import (
	"testing"

	. "github.com/octohelm/x/testing/v2"
)

func TestParseStringNormalizesFormatting(t *testing.T) {
	ex, _ := ParseString(`
select(
	when(
		pipe(get('x'), eq(1.1)),
		eq(1),
	),
	eq(2),
)
`)
	Then(t, "表达式字符串会被解析并输出稳定格式",
		Expect(Stringify(ex), Equal(`select(when(pipe(get("x"),eq(1.1)),eq(1)),eq(2))`)),
	)
}
