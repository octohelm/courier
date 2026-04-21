package raw

import (
	"regexp"
	"testing"

	. "github.com/octohelm/x/testing/v2"
)

func TestCompareBranches(t0 *testing.T) {
	Then(t0, "compare 覆盖字符串与错误分支",
		ExpectMust(func() error {
			r, err := Compare(StringValue("a"), StringValue("b"))
			if err != nil || r != -1 {
				return errRaw("unexpected string compare result")
			}
			r, err = Compare(StringValue("b"), StringValue("a"))
			if err != nil || r != 1 {
				return errRaw("unexpected string compare result")
			}
			r, err = Compare(StringValue("a"), StringValue("a"))
			if err != nil || r != 0 {
				return errRaw("unexpected string compare result")
			}
			return nil
		}),
		ExpectDo(func() error {
			_, err := Compare(BoolValue(true), StringValue("x"))
			return err
		}, ErrorMatch(regexp.MustCompile("not comparable"))),
	)
}

func TestMathErrorAndMixedBranches(t0 *testing.T) {
	Then(t0, "math helper 覆盖混合类型与错误分支",
		ExpectMust(func() error {
			if v, err := Add(FloatValue(1.5), IntValue(2)); err != nil || v.(float64) != 3.5 {
				return errRaw("unexpected add float/int")
			}
			if v, err := Sub(FloatValue(0.2), FloatValue(1.2)); err != nil || v.(float64) != 1 {
				return errRaw("unexpected sub float")
			}
			if v, err := Mul(UintValue(2), IntValue(3)); err != nil || v.(int64) != 6 {
				return errRaw("unexpected mul uint/int")
			}
			if v, err := Div(IntValue(2), IntValue(4)); err != nil || v.(int64) != 2 {
				return errRaw("unexpected int div")
			}
			if v, err := Div(IntValue(2), IntValue(3)); err != nil || v.(float64) != 1.5 {
				return errRaw("unexpected float div")
			}
			if v, err := Div(UintValue(2), UintValue(3)); err != nil || v.(float64) != 1.5 {
				return errRaw("unexpected uint div")
			}
			if v, err := Pow(UintValue(3), UintValue(2)); err != nil || v.(uint64) != 8 {
				return errRaw("unexpected uint pow")
			}
			if v, err := Pow(IntValue(3), IntValue(2)); err != nil || v.(int64) != 8 {
				return errRaw("unexpected int pow")
			}
			return nil
		}),
		ExpectDo(func() error {
			_, err := Add(StringValue("x"), IntValue(1))
			return err
		}, ErrorMatch(regexp.MustCompile("can't add"))),
		ExpectDo(func() error {
			_, err := Sub(StringValue("x"), IntValue(1))
			return err
		}, ErrorMatch(regexp.MustCompile("can't minus"))),
		ExpectDo(func() error {
			_, err := Mul(StringValue("x"), IntValue(1))
			return err
		}, ErrorMatch(regexp.MustCompile("can't multiply"))),
		ExpectDo(func() error {
			_, err := Div(IntValue(0), IntValue(1))
			return err
		}, ErrorMatch(regexp.MustCompile("can't divide 0"))),
		ExpectDo(func() error {
			_, err := Pow(StringValue("x"), IntValue(1))
			return err
		}, ErrorMatch(regexp.MustCompile("can't pow"))),
	)
}
