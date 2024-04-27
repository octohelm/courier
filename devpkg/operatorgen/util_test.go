package operatorgen

import (
	"fmt"
	"testing"
)

func Test_normalizeFormat(t *testing.T) {
	cases := []struct {
		actual string
		expect string
	}{
		{
			"%s %q %v %.1f %q %%",
			"%v %v %v %v %v %%",
		},
		{
			"%[2]s %[1]q",
			"%v %v",
		},
	}

	for _, tc := range cases {
		t.Run(fmt.Sprintf("%q", tc.actual), func(t *testing.T) {
			if expect := normalizeFormat(tc.actual); expect != tc.expect {
				t.Fatalf("expect: %q, actual: %q", tc.expect, tc.actual)
			}
		})
	}
}
