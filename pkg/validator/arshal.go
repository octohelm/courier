package validator

import (
	"bytes"
	"io"

	"github.com/go-json-experiment/json"
	"github.com/go-json-experiment/json/jsontext"
	jsonv1 "github.com/go-json-experiment/json/v1"
	"github.com/octohelm/courier/pkg/validator/internal"
)

func UnmarshalDecode(dec *jsontext.Decoder, out any, options ...jsontext.Options) error {
	return internal.UnmarshalDecode(dec, out, options...)
}

func UnmarshalRead(r io.Reader, out any, options ...jsontext.Options) error {
	return internal.UnmarshalDecode(jsontext.NewDecoder(r), out, options...)
}

func Unmarshal(in []byte, out any, options ...jsontext.Options) error {
	return internal.UnmarshalDecode(jsontext.NewDecoder(bytes.NewBuffer(in)), out, options...)
}

func MarshalWrite(w io.Writer, out any, options ...jsontext.Options) error {
	return json.MarshalWrite(w, out, append(options, jsonv1.OmitEmptyWithLegacySemantics(true))...)
}

func MarshalEncode(enc *jsontext.Encoder, out any, options ...jsontext.Options) error {
	return json.MarshalEncode(enc, out, append(options, jsonv1.OmitEmptyWithLegacySemantics(true))...)
}

func Marshal(out any, options ...jsontext.Options) ([]byte, error) {
	return json.Marshal(out, append(options, jsonv1.OmitEmptyWithLegacySemantics(true))...)
}
