package validator

import (
	"bytes"
	"io"

	"github.com/go-json-experiment/json"
	"github.com/go-json-experiment/json/jsontext"
	jsonv1 "github.com/go-json-experiment/json/v1"

	"github.com/octohelm/courier/pkg/validator/internal"
)

// UnmarshalDecode 从JSON解码器反序列化数据。
// Unmarshal 从字节数组反序列化JSON数据。
func UnmarshalDecode(dec *jsontext.Decoder, out any, options ...jsontext.Options) error {
	return internal.UnmarshalDecode(dec, out, options...)
}

// UnmarshalRead 从读取器反序列化JSON数据。
func UnmarshalRead(r io.Reader, out any, options ...jsontext.Options) error {
	return internal.UnmarshalDecode(jsontext.NewDecoder(r), out, options...)
}

// Unmarshal 从字节数组反序列化JSON数据。
func Unmarshal(in []byte, out any, options ...jsontext.Options) error {
	return internal.UnmarshalDecode(jsontext.NewDecoder(bytes.NewBuffer(in)), out, options...)
}

// MarshalWrite 将数据序列化为JSON并写入写入器。
func MarshalWrite(w io.Writer, out any, options ...jsontext.Options) error {
	return json.MarshalWrite(w, out, append(options, jsonv1.OmitEmptyWithLegacySemantics(true))...)
}

// MarshalEncode 将数据序列化为JSON并写入编码器。
func MarshalEncode(enc *jsontext.Encoder, out any, options ...jsontext.Options) error {
	return json.MarshalEncode(enc, out, append(options, jsonv1.OmitEmptyWithLegacySemantics(true))...)
}

// Marshal 将数据序列化为JSON字节数组。
func Marshal(out any, options ...jsontext.Options) ([]byte, error) {
	return json.Marshal(out, append(options, jsonv1.OmitEmptyWithLegacySemantics(true))...)
}
