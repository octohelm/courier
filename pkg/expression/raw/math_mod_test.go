package raw

import (
	"fmt"
	testingx "github.com/octohelm/x/testing"
	"testing"
)

var modCases = [][]any{
	{2, 3, int64(1)},
	{2, uint(3), int64(1)},
	{uint(2), uint(3), uint64(1)},
	{2, float64(3), float64(1)},
	{2.2, 4.5, 0.1},
}

func TestMod(t *testing.T) {
	for _, c := range modCases {
		t.Run(fmt.Sprintf("%T(%v) mod %T(%v) = %T(%v)", c[1], c[1], c[0], c[0], c[2], c[2]), func(t *testing.T) {
			v, err := Mod(ValueOf(c[0]), ValueOf(c[1]))

			testingx.Expect(t, err, testingx.Be[error](nil))
			testingx.Expect(t, v, testingx.Equal(c[2]))
		})
	}
}

func BenchmarkMod(b *testing.B) {
	for _, c := range modCases {
		b.Run(fmt.Sprintf("%T(%v) mod %T(%v)", c[1], c[1], c[0], c[0]), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_, _ = Mod(ValueOf(c[0]), ValueOf(c[1]))
			}
		})
	}
}
