package raw

import (
	"fmt"
	"testing"

	"github.com/onsi/gomega"
)

var addCases = [][]any{
	{1, 1, int64(2)},
	{1, uint(1), int64(2)},
	{uint(1), uint(1), uint64(2)},
	{1, float64(1), float64(2)},
}

func TestAdd(t *testing.T) {
	for _, c := range addCases {
		t.Run(fmt.Sprintf("%T(%v) + %T(%v) = %T(%v)", c[0], c[0], c[1], c[1], c[2], c[2]), func(t *testing.T) {
			v, err := Add(ValueOf(c[0]), ValueOf(c[1]))
			gomega.NewWithT(t).Expect(err).To(gomega.BeNil())
			gomega.NewWithT(t).Expect(v).To(gomega.Equal(c[2]))
		})
	}
}

func BenchmarkAdd(b *testing.B) {
	for _, c := range addCases {
		b.Run(fmt.Sprintf("%T(%v) + %T(%v)", c[1], c[1], c[0], c[0]), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_, _ = Add(ValueOf(c[0]), ValueOf(c[1]))
			}
		})
	}
}
