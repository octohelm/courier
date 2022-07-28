package raw

import (
	"context"
	"fmt"
	"strconv"
	"testing"
)

func BenchmarkMapValue_Iter(b *testing.B) {
	m := MapValue{}
	for i := 1; i < 10_0000; i++ {
		m[strconv.Itoa(i)] = ValueOf(i)
	}

	b.Run("iter", func(b *testing.B) {
		for item := range m.Iter(context.Background()) {
			_ = item
		}
	})

	b.Run("direct", func(b *testing.B) {
		for k := range m {
			_ = &entity{
				key:   StringValue(k),
				value: m[k],
			}
		}
	})
}

func TestMapValue(t *testing.T) {
	m := MapValue{}
	for i := 1; i < 10000; i++ {
		m[strconv.Itoa(i)] = ValueOf(i)
	}

	t.Run("iter", func(t *testing.T) {
		count := 0

		ctx, cancel := context.WithCancel(context.Background())
		for item := range m.Iter(ctx) {
			if count > 100 {
				break
			}
			_ = item
			count++
		}
		cancel()
		fmt.Println("done", count)
	})
}
