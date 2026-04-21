package raw

import (
	"regexp"
	"testing"

	. "github.com/octohelm/x/testing/v2"
)

func TestAdditionalMathAndValueCoverage(t *testing.T) {
	Then(t, "补齐 raw 数学与取值分支",
		ExpectMust(func() error {
			if v, err := Mod(IntValue(5), IntValue(2)); err != nil || v.(int64) != 2 {
				return errRaw("unexpected int mod")
			}
			if v, err := Mod(UintValue(5), UintValue(2)); err != nil || v.(uint64) != 2 {
				return errRaw("unexpected uint mod")
			}
			if v, err := Compare(FloatValue(1.5), FloatValue(1.5)); err != nil || v != 0 {
				return errRaw("unexpected float compare")
			}
			if v := ValueOf(IntValue(1)); v == nil || v.Kind() != Int {
				return errRaw("unexpected value passthrough kind")
			}
			if v := ValueOf(map[string]any{"a": 1}); v == nil || v.Kind() != Map {
				return errRaw("unexpected any map value kind")
			}
			return nil
		}),
		ExpectDo(func() error {
			_, err := Mod(StringValue("x"), IntValue(1))
			return err
		}, ErrorMatch(regexp.MustCompile("can't mod"))),
	)
}
