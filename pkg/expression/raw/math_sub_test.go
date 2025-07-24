package raw

import (
	"fmt"
	"testing"

	"github.com/octohelm/x/testing/bdd"
)

var subCases = [][]any{
	{1, 2, int64(1)},
	{1, uint(2), int64(1)},
	{uint(1), uint(2), uint64(1)},
	{1, float64(2), float64(1)},
}

func TestSub(t *testing.T) {
	for _, c := range subCases {
		bdd.FromT(t).When(fmt.Sprintf("%T(%v) - %T(%v)  = %T(%v)", c[1], c[1], c[0], c[0], c[2], c[2]), func(b bdd.T) {
			v, err := Sub(ValueOf(c[0]), ValueOf(c[1]))

			b.Then("success",
				bdd.NoError(err),
				bdd.Equal(c[2], v),
			)
		})
	}
}

func BenchmarkSub(b *testing.B) {
	for _, c := range subCases {
		b.Run(fmt.Sprintf("%T(%v) - %T(%v)", c[1], c[1], c[0], c[0]), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_, _ = Sub(ValueOf(c[0]), ValueOf(c[1]))
			}
		})
	}
}
