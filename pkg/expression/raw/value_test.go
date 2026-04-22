package raw

import (
	"reflect"
	"testing"

	. "github.com/octohelm/x/testing/v2"
)

type textValue string

func (v textValue) MarshalText() ([]byte, error) {
	return []byte(v), nil
}

func TestValueAndConvertHelpers(t0 *testing.T) {
	Then(t0, "raw value 与转换辅助方法覆盖常见分支",
		Expect(ValueOf(textValue("x")).Kind(), Equal(String)),
		Expect(ValueOf(float32(1)).Kind(), Equal(Float)),
		Expect(ValueOf(int8(1)).Kind(), Equal(Int)),
		Expect(ValueOf(uint8(1)).Kind(), Equal(Uint)),
		Expect(ValueOf("x").Kind(), Equal(String)),
		Expect(ValueOf(true).Kind(), Equal(Bool)),
		Expect(ValueOf([]any{1, "x"}).Kind(), Equal(Array)),
		Expect(ValueOf(map[string]any{"x": 1}).Kind(), Equal(Map)),
		Expect(ValueOf(struct{}{}), Equal[Value](nil)),
		Expect(ArrayValue{IntValue(1), StringValue("x")}.Kind(), Equal(Array)),
		Expect(ArrayValue{IntValue(1), StringValue("x")}.Len(), Equal(2)),
		Expect(ArrayValue{IntValue(1), StringValue("x")}.Index(1), Equal[Value](StringValue("x"))),
		Expect(MapValue{"x": IntValue(1)}.Kind(), Equal(Map)),
		Expect(FloatValue(1).Kind(), Equal(Float)),
		Expect(IntValue(1).Kind(), Equal(Int)),
		Expect(UintValue(1).Kind(), Equal(Uint)),
		Expect(StringValue("x").Kind(), Equal(String)),
		Expect(BoolValue(true).Kind(), Equal(Bool)),
		Expect(Len(StringValue("abc")), Equal(3)),
		Expect(Len(ArrayValue{IntValue(1)}), Equal(1)),
		Expect(Len(MapValue{"x": IntValue(1)}), Equal(1)),
		Expect(Len(BoolValue(true)), Equal(0)),
		Expect(ToString(StringValue("abc")), Equal("abc")),
		Expect(ToString(FloatValue(1.5)), Equal("1.5")),
		Expect(ToString(IntValue(2)), Equal("2")),
		Expect(ToString(UintValue(3)), Equal("3")),
		Expect(ToString(BoolValue(true)), Equal("true")),
		Expect(ToString(BoolValue(false)), Equal("false")),
		Expect(ToString(ArrayValue{}), Equal("")),
		Expect(ToFloat(FloatValue(1.5)), Equal(1.5)),
		Expect(ToFloat(IntValue(2)), Equal(2.0)),
		Expect(ToFloat(UintValue(3)), Equal(3.0)),
		Expect(ToFloat(BoolValue(true)), Equal(1.0)),
		Expect(ToFloat(BoolValue(false)), Equal(0.0)),
		Expect(ToInt(FloatValue(1.5)), Equal(int64(1))),
		Expect(ToInt(IntValue(2)), Equal(int64(2))),
		Expect(ToInt(UintValue(3)), Equal(int64(3))),
		Expect(ToInt(BoolValue(true)), Equal(int64(1))),
		Expect(ToInt(BoolValue(false)), Equal(int64(0))),
		Expect(ToUint(FloatValue(1.5)), Equal(uint64(1))),
		Expect(ToUint(IntValue(2)), Equal(uint64(2))),
		Expect(ToUint(UintValue(3)), Equal(uint64(3))),
		Expect(ToUint(BoolValue(true)), Equal(uint64(1))),
		Expect(ToUint(BoolValue(false)), Equal(uint64(0))),
		Expect(ToBool(FloatValue(1)), Equal(true)),
		Expect(ToBool(FloatValue(0)), Equal(false)),
		Expect(ToBool(IntValue(1)), Equal(true)),
		Expect(ToBool(IntValue(0)), Equal(false)),
		Expect(ToBool(UintValue(1)), Equal(true)),
		Expect(ToBool(UintValue(0)), Equal(false)),
		Expect(ToBool(BoolValue(true)), Equal(true)),
		Expect(ToBool(BoolValue(false)), Equal(false)),
	)
}

func TestIterHelpers(t0 *testing.T) {
	Then(t0, "array/map iterator 与 iter value 行为正确",
		ExpectMust(func() error {
			arr := ArrayValue{IntValue(1), StringValue("x")}
			iter := arr.Iter()
			got := make([][2]Value, 0)
			for iter.Next() {
				v := iter.Val()
				got = append(got, [2]Value{v.Key(), v.Value()})
			}
			if !reflect.DeepEqual(got, [][2]Value{
				{IntValue(0), IntValue(1)},
				{IntValue(1), StringValue("x")},
			}) {
				return errRaw("unexpected array iter result")
			}
			return nil
		}),
		ExpectMust(func() error {
			m := MapValue{"x": IntValue(1)}
			keys := m.Keys()
			if len(keys) != 1 || keys[0] != "x" {
				return errRaw("unexpected map keys")
			}
			iter := m.Iter()
			if !iter.Next() {
				return errRaw("expected map iter next")
			}
			v := iter.Val()
			if v.Key() != StringValue("x") || v.Value() != IntValue(1) {
				return errRaw("unexpected map iter value")
			}
			return nil
		}),
	)
}

func errRaw(msg string) error {
	return &rawErr{msg: msg}
}

type rawErr struct{ msg string }

func (e *rawErr) Error() string { return e.msg }
