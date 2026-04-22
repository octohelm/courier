package rules

import (
	"fmt"
	"testing"

	. "github.com/octohelm/x/testing/v2"
)

func TestSlashAndUnslash(t *testing.T) {
	cases := [][]string{
		{`/\w+\/test/`, `\w+/test`},
		{`/a/`, `a`},
		{`/abc/`, `abc`},
		{`/☺/`, `☺`},
		{`/\xFF/`, `\xFF`},
		{`/\377/`, `\377`},
		{`/\u1234/`, `\u1234`},
		{`/\U00010111/`, `\U00010111`},
		{`/\U0001011111/`, `\U0001011111`},
		{`/\a\b\f\n\r\t\v\\\"/`, `\a\b\f\n\r\t\v\\\"`},
		{`/\//`, `/`},
	}

	for i := range cases {
		c := cases[i]

		t.Run("unslash "+c[0], func(t *testing.T) {
			Then(t, "Unslash 会还原原始文本", ExpectMust(func() error {
				r, err := Unslash([]byte(c[0]))
				if err != nil {
					return err
				}
				if string(r) != c[1] {
					return fmt.Errorf("unexpected unslash result: %s", r)
				}
				return nil
			}))
		})

		t.Run("slash "+c[1], func(t *testing.T) {
			Then(t, "Slash 会生成可回放的正则文本", ExpectMust(func() error {
				v := Slash([]byte(c[1]))
				if string(v) != c[0] {
					return fmt.Errorf("unexpected slash result: %s", v)
				}
				return nil
			}))
		})
	}

	casesForFailed := [][]string{
		{`/`, ``},
		{`/adfadf`, ``},
	}

	for i := range casesForFailed {
		c := casesForFailed[i]

		t.Run("unslash invalid "+c[0], func(t *testing.T) {
			Then(t, "非法输入会返回错误",
				ExpectMust(func() error {
					_, err := Unslash([]byte(c[0]))
					if err == nil {
						return fmt.Errorf("expected unslash error")
					}
					return nil
				}),
			)
		})
	}
}
