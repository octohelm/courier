package raw

import (
	"fmt"
	"testing"

	. "github.com/octohelm/x/testing/v2"
)

var divideCases = [][]any{
	{2, 8, int64(4)},
	{2, uint(8), int64(4)},
	{uint(2), uint(8), uint64(4)},
	{2, float64(8), float64(4)},
	{2, float64(7), 3.5},
	{uint8(2), uint64(7), 3.5},
}

func TestDiv(t *testing.T) {
	for _, c := range divideCases {
		t.Run(fmt.Sprintf("%T(%v)/%T(%v)", c[0], c[0], c[1], c[1]), func(t *testing.T) {
			Then(t, "除法会返回预期结果", ExpectMust(func() error {
				v, err := Div(ValueOf(c[0]), ValueOf(c[1]))
				if err != nil {
					return err
				}
				if v != c[2] {
					return errRaw("unexpected divide result")
				}
				return nil
			}))
		})
	}
}

func BenchmarkDiv(b *testing.B) {
	for _, c := range divideCases {
		b.Run(fmt.Sprintf("%T(%v) / %T(%v)", c[1], c[1], c[0], c[0]), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_, _ = Div(ValueOf(c[0]), ValueOf(c[1]))
			}
		})
	}
}
