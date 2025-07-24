package rules

import (
	"testing"

	"github.com/octohelm/x/testing/bdd"
)

func TestSlashUnslash(t *testing.T) {
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

	b := bdd.FromT(t)

	for i := range cases {
		c := cases[i]

		b.When("unslash:"+c[0], func(b bdd.T) {
			r, err := Unslash([]byte(c[0]))

			b.Then("success",
				bdd.NoError(err),
				bdd.Equal(c[1], string(r)),
			)
		})

		b.When("slash:"+c[1], func(b bdd.T) {
			v := Slash([]byte(c[1]))

			b.Then("success",
				bdd.Equal(c[0], string(v)),
			)
		})
	}

	casesForFailed := [][]string{
		{`/`, ``},
		{`/adfadf`, ``},
	}

	for i := range casesForFailed {
		c := casesForFailed[i]

		b.When("unslash:"+c[0], func(b bdd.T) {
			_, err := Unslash([]byte(c[0]))
			b.Then("success",
				bdd.HasError(err),
			)
		})
	}
}
