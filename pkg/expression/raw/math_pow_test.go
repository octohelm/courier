package raw

import (
	"fmt"
	"testing"

	. "github.com/octohelm/x/testing/v2"
)

var powCases = [][]any{
	{2, 2, int64(4)},
	{2, uint(2), int64(4)},
	{uint(2), uint(2), uint64(4)},
	{2, float64(2), float64(4)},
	{2, 2.2, 4.84},
}

func TestPow(t *testing.T) {
	for _, c := range powCases {
		t.Run(fmt.Sprintf("%T(%v)^%T(%v)", c[0], c[0], c[1], c[1]), func(t *testing.T) {
			Then(t, "幂运算会返回预期结果", ExpectMust(func() error {
				v, err := Pow(ValueOf(c[0]), ValueOf(c[1]))
				if err != nil {
					return err
				}
				if v != c[2] {
					return errRaw("unexpected pow result")
				}
				return nil
			}))
		})
	}
}

func BenchmarkPow(b *testing.B) {
	for _, c := range powCases {
		b.Run(fmt.Sprintf("%T(%v) pow %T(%v)", c[1], c[1], c[0], c[0]), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_, _ = Pow(ValueOf(c[0]), ValueOf(c[1]))
			}
		})
	}
}
