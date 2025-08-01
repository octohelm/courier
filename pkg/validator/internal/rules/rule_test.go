package rules

import (
	"testing"

	"github.com/octohelm/x/testing/bdd"
)

func TestParseRule(t *testing.T) {
	cases := [][]string{
		// simple
		{`@email`, `@email`},

		// with parameters
		{`@map<@email,         @url>`, `@map<@email,@url>`},
		{`@map<@string,>`, `@map<@string,>`},
		{`@map<,@string>`, `@map<,@string>`},
		{`@float32<10,6>`, `@float32<10,6>`},
		{`@float32<10,-1>`, `@float32<10,-1>`},
		{`@slice<@string>`, `@slice<@string>`},

		// with range
		{`@slice[0,   10]`, `@slice[0,10]`},
		{`@array[10]`, `@array[10]`},
		{`@string[0,)`, `@string[0,)`},
		{`@string[0,)`, `@string[0,)`},
		{`@int(0,)`, `@int(0,)`},
		{`@int(,1)`, `@int(,1)`},
		{`@float32(1.10,)`, `@float32(1.10,)`},

		// with values
		{`@string{A, B,    C}`, `@string{A,B,C}`},
		{`@string{, B,    C}`, `@string{,B,C}`},
		{`@uint{%2}`, `@uint{%2}`},

		// with value matrix
		{`@string{A, B,    C}{a,b}`, `@string{A,B,C}{a,b}`},

		// with not required mark or default value
		{`@string?`, `@string?`},
		{`@string ?`, `@string?`},
		{`@string = `, `@string = ''`},
		{`@string = '\''`, `@string = '\''`},
		{`@string = 'default value'`, `@string = 'default value'`},
		{`@string = 'defa\'ult\ value'`, `@string = 'defa\'ult\ value'`},
		{`@string = 13123`, `@string = '13123'`},
		{`@string = 1.1`, `@string = '1.1'`},

		// with regexp
		{`@string[0,10]/\w+/`, `@string[0,10]/\w+/`},
		{`@string/\w+/`, `@string/\w+/`},
		{`@string/\w+     $/`, `@string/\w+     $/`},
		{`@string/\w+\/abc/`, `@string/\w+\/abc/`},
		{`@string/\w+\/\/abc/`, `@string/\w+\/\/abc/`},
		{`@string/^\w+\/test/`, `@string/^\w+\/test/`},

		// custom format
		{"@email", "@email"},
		{"@uuid7", "@uuid7"},
		{"@uri-reference", "@uri-reference"},

		// composes
		{`@string = 's'/\w+/`, `@string/\w+/ = 's'`},
		{`@map<,@string[1,]>`, `@map<,@string[1,]>`},
		{`@map<@string,>[1,2]`, `@map<@string,>[1,2]`},
		{`@map<@string = 's',>[1,2]`, `@map<@string = 's',>[1,2]`},
		{`@slice<@float64<10,4>[-1.000,100.000]?>`, `@slice<@float64<10,4>[-1.000,100.000]?>`},
	}

	b := bdd.FromT(t)

	for i := range cases {
		c := cases[i]

		b.When("parse rule:"+c[0], func(b bdd.T) {
			r, err := ParseRuleString(c[0])

			b.Then("success",
				bdd.NoError(err),
				bdd.Equal(c[1], string(r.Bytes())),
			)
		})
	}
}

func TestParseRuleFailed(t *testing.T) {
	cases := []string{
		`@`,
		`@name<`,
		`@name[`,
		`@name(`,
		`@name{`,
		`@name/`,
		`@name)`,
		`@name<@sub[>`,
		`@name</>`,
		`@/`,
		`@name?=`,
	}

	for _, c := range cases {
		_, err := ParseRuleString(c)
		t.Logf("%s %s", c, err)
	}
}

func TestRule(t *testing.T) {
	bdd.FromT(t).When("parse", func(b bdd.T) {
		_, err := ParseRuleString("@string{A,B,C}{a,b}{1,2}")

		b.Then("success",
			bdd.NoError(err),
		)
	})
}
