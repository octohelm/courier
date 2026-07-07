package v1

import (
	"encoding/json"
	"strings"
	"testing"

	. "github.com/octohelm/x/testing/v2"
)

func TestUnion_IsZero(t *testing.T) {
	u := Union{}
	Then(t, "零值应为 true", Expect(u.IsZero(), Equal(true)))
}

func TestUnion_SetUnderlying(t *testing.T) {
	u := Union{}
	u.SetUnderlying(&StringKinded{})

	Then(t, "SetUnderlying 后应有值",
		Expect(u.IsZero(), Equal(false)),
	)
	Then(t, "对应字段应被设置",
		Expect(u.String != nil, Equal(true)),
	)
	Then(t, "其他字段应为 nil",
		Expect(u.Bool, Equal[*BoolKinded](nil)),
	)
}

func TestUnion_IsZero_Normal(t *testing.T) {
	u := Union{
		String: &StringKinded{},
	}
	Then(t, "非零值应为 false", Expect(u.IsZero(), Equal(false)))
}

func TestUnion_MarshalJSON_Zero(t *testing.T) {
	u := Union{}

	data, err := json.Marshal(&u)
	Then(t, "零值序列化应无错误", Expect(err, Equal[error](nil)))
	Then(t, "零值应输出空对象", Expect(string(data), Equal("{}")))
}

func TestUnion_MarshalJSON_StringKinded(t *testing.T) {
	u := Union{
		String: &StringKinded{
			Kind:  "String",
			Value: "hello",
		},
	}

	data, err := json.Marshal(&u)
	Then(t, "序列化应无错误", Expect(err, Equal[error](nil)))

	s := string(data)
	Then(t, "输出应包含 discriminator", Expect(strings.Contains(s, `"kind":"String"`), Equal(true)))
	Then(t, "输出应包含 value", Expect(strings.Contains(s, `"value":"hello"`), Equal(true)))
}

func TestUnion_MarshalJSON_BoolKinded(t *testing.T) {
	u := Union{
		Bool: &BoolKinded{
			Kind:  "Bool",
			Value: true,
		},
	}

	data, err := json.Marshal(&u)
	Then(t, "序列化应无错误", Expect(err, Equal[error](nil)))

	s := string(data)
	Then(t, "输出应包含 discriminator", Expect(strings.Contains(s, `"kind":"Bool"`), Equal(true)))
	Then(t, "输出应包含 value", Expect(strings.Contains(s, `"value":true`), Equal(true)))
}

func TestUnion_UnmarshalJSON_StringKinded(t *testing.T) {
	data := []byte(`{"kind":"String","value":"hello"}`)

	u := Union{}
	err := json.Unmarshal(data, &u)
	Then(t, "反序列化应无错误", Expect(err, Equal[error](nil)))
	Then(t, "应正确设置 String 字段",
		Expect(u.String != nil, Equal(true)),
	)
	Then(t, "String 字段的 value 应正确",
		Expect(u.String.Value, Equal("hello")),
	)
	Then(t, "Bool 字段应保持为 nil",
		Expect(u.Bool, Equal[*BoolKinded](nil)),
	)
}

func TestUnion_UnmarshalJSON_BoolKinded(t *testing.T) {
	data := []byte(`{"kind":"Bool","value":true}`)

	u := Union{}
	err := json.Unmarshal(data, &u)
	Then(t, "反序列化应无错误", Expect(err, Equal[error](nil)))
	Then(t, "应正确设置 Bool 字段",
		Expect(u.Bool != nil, Equal(true)),
	)
	Then(t, "Bool 字段的 value 应正确",
		Expect(u.Bool.Value, Equal(true)),
	)
	Then(t, "String 字段应保持为 nil",
		Expect(u.String, Equal[*StringKinded](nil)),
	)
}

func TestUnion_UnmarshalJSON_Null(t *testing.T) {
	data := []byte(`null`)

	u := Union{}
	err := json.Unmarshal(data, &u)
	Then(t, "null 反序列化应无错误", Expect(err, Equal[error](nil)))
	Then(t, "零值应为 true", Expect(u.IsZero(), Equal(true)))
}

func TestUnion_UnmarshalJSON_EmptyObject(t *testing.T) {
	data := []byte(`{}`)

	u := Union{}
	err := json.Unmarshal(data, &u)
	Then(t, "空对象反序列化应无错误", Expect(err, Equal[error](nil)))
	Then(t, "零值应为 true", Expect(u.IsZero(), Equal(true)))
}

func TestUnion_RoundTrip(t *testing.T) {
	u := Union{
		String: &StringKinded{
			Kind:  "String",
			Value: "roundtrip",
		},
	}

	data, err := json.Marshal(&u)
	Then(t, "序列化应无错误", Expect(err, Equal[error](nil)))

	u2 := Union{}
	err = json.Unmarshal(data, &u2)
	Then(t, "反序列化应无错误", Expect(err, Equal[error](nil)))
	Then(t, "String 字段应正确恢复",
		Expect(u2.String != nil, Equal(true)),
	)
	Then(t, "字段值应一致",
		Expect(u2.String.Value, Equal("roundtrip")),
	)
}
