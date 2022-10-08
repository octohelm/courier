package raw

import (
	"fmt"
	"testing"

	"github.com/onsi/gomega"
)

var multiplyCases = [][]any{
	{2, 2, int64(4)},
	{2, uint(2), int64(4)},
	{uint(2), uint(2), uint64(4)},
	{2, float64(2), float64(4)},
}

func TestMultiply(t *testing.T) {
	for _, c := range multiplyCases {
		t.Run(fmt.Sprintf("%T(%v) * %T(%v)  = %T(%v)", c[1], c[1], c[0], c[0], c[2], c[2]), func(t *testing.T) {
			v, err := Mul(ValueOf(c[0]), ValueOf(c[1]))
			gomega.NewWithT(t).Expect(err).To(gomega.BeNil())
			gomega.NewWithT(t).Expect(v).To(gomega.Equal(c[2]))
		})
	}
}

func BenchmarkMultiply(b *testing.B) {
	for _, c := range multiplyCases {
		b.Run(fmt.Sprintf("%T(%v) * %T(%v)", c[1], c[1], c[0], c[0]), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_, _ = Mul(ValueOf(c[0]), ValueOf(c[1]))
			}
		})
	}
}
