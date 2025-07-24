package internal

import (
	"testing"
)

func TestMaxIntAndMinInt(t *testing.T) {
	cases := [][]int64{
		{MinInt(8), -128},
		{MaxInt(8), 127},
		{MinInt(16), -32768},
		{MaxInt(16), 32767},
		{MinInt(32), -2147483648},
		{MaxInt(32), 2147483647},
		{MinInt(64), -9223372036854775808},
		{MaxInt(64), 9223372036854775807},
	}
	for _, values := range cases {
		if values[0] != values[1] {
			t.Errorf("got %d, want %d", values[0], values[1])
		}
	}
}
