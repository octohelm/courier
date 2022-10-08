package raw

import (
	"fmt"
	"testing"

	"github.com/onsi/gomega"
)

var compareCases = map[int][][]any{
	0: {
		{1, 1},
		{1, uint(1)},
		{1, float64(1)},
		{"a", "a"},
	},
	-1: {
		{1, 2},
		{1, uint(2)},
		{1, float64(2)},
		{1, 2},
		{uint(1), float64(2)},
		{float64(1), 2},
		{"a", "b"},
	},
	1: {
		{3, 2},
		{3, uint(2)},
		{3, float64(2)},
		{3, 2},
		{uint(3), float64(2)},
		{float64(3), 2},
	},
}

func TestCompare(t *testing.T) {
	for i, cs := range compareCases {
		switch i {
		case -1:
			for _, c := range cs {
				t.Run(fmt.Sprintf("%T(%v) should less than %T(%v)", c[0], c[0], c[1], c[1]), func(t *testing.T) {
					v, err := Compare(ValueOf(c[0]), ValueOf(c[1]))
					gomega.NewWithT(t).Expect(err).To(gomega.BeNil())
					gomega.NewWithT(t).Expect(v).To(gomega.Equal(i))
				})
			}
		case 1:
			for _, c := range cs {
				t.Run(fmt.Sprintf("%T(%v) should great than %T(%v)", c[0], c[0], c[1], c[1]), func(t *testing.T) {
					v, err := Compare(ValueOf(c[0]), ValueOf(c[1]))
					gomega.NewWithT(t).Expect(err).To(gomega.BeNil())
					gomega.NewWithT(t).Expect(v).To(gomega.Equal(i))
				})
			}
		case 0:
			for _, c := range cs {
				t.Run(fmt.Sprintf("%T(%v) should equal %T(%v)", c[0], c[0], c[1], c[1]), func(t *testing.T) {
					v, err := Compare(ValueOf(c[0]), ValueOf(c[1]))
					gomega.NewWithT(t).Expect(err).To(gomega.BeNil())
					gomega.NewWithT(t).Expect(v).To(gomega.Equal(i))
				})
			}
		}
	}
}

func BenchmarkCompare(b *testing.B) {
	for _, cs := range compareCases {
		for _, c := range cs {
			b.Run(fmt.Sprintf("compare %T(%v),%T(%v)", c[0], c[0], c[1], c[1]), func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					_, _ = Compare(ValueOf(c[0]), ValueOf(c[1]))
				}
			})
		}
	}
}
