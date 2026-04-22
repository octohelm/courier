package transformers_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"testing"

	. "github.com/octohelm/x/testing/v2"

	contentinternal "github.com/octohelm/courier/pkg/content/internal"
)

type directJSON struct {
	Value string
}

type stringAlias string

func (v directJSON) MarshalJSON() ([]byte, error) {
	return []byte(`{"value":"` + v.Value + `"}`), nil
}

type failingJSON struct{}

func (f failingJSON) MarshalJSON() ([]byte, error) {
	return nil, errors.New("marshal failed")
}

type octetMetaTarget struct {
	Name string
	Type string
	Data []byte
}

func (t *octetMetaTarget) SetFilename(name string)  { t.Name = name }
func (t *octetMetaTarget) SetContentType(ct string) { t.Type = ct }
func (t *octetMetaTarget) ReadFromCloser(r io.ReadCloser) (int64, error) {
	defer r.Close()
	data, err := io.ReadAll(r)
	t.Data = data
	return int64(len(data)), err
}

func TestTransformerMediaTypesAndFallbacks(t *testing.T) {
	Then(t, "各 transformer 可通过 internal.New 暴露稳定 MediaType", ExpectMust(func() error {
		cases := []struct {
			typ       reflect.Type
			mediaType string
			action    string
			expect    string
		}{
			{typ: reflect.TypeFor[string](), action: "marshal", expect: "text/plain"},
			{typ: reflect.TypeFor[[]byte](), action: "marshal", expect: "application/octet-stream"},
			{typ: reflect.TypeFor[struct{ Name string }](), mediaType: "json", action: "marshal", expect: "application/json"},
			{typ: reflect.TypeFor[struct{ Name string }](), mediaType: "urlencoded", action: "marshal", expect: "application/x-www-form-urlencoded"},
			{typ: reflect.TypeFor[struct{ Name string }](), mediaType: "multipart", action: "marshal", expect: "multipart/form-data"},
		}

		for _, c := range cases {
			tr, err := contentinternal.New(c.typ, c.mediaType, c.action)
			if err != nil {
				return err
			}
			if tr.MediaType() != c.expect {
				return errTransformer("unexpected media type: " + tr.MediaType())
			}
		}
		return nil
	}))
}

func TestJSONTransformerDirectBranches(t *testing.T) {
	Then(t, "json transformer 覆盖 direct marshal 与 direct unmarshal 分支", ExpectMust(func() error {
		tr, err := contentinternal.New(reflect.TypeFor[directJSON](), "json", "marshal")
		if err != nil {
			return err
		}
		content, err := tr.Prepare(context.Background(), directJSON{Value: "demo"})
		if err != nil {
			return err
		}
		defer content.Close()
		data, err := io.ReadAll(content)
		if err != nil {
			return err
		}
		if string(data) != `{"value":"demo"}` {
			return errTransformer("unexpected direct marshal output")
		}

		unmarshalTr, err := contentinternal.New(reflect.TypeFor[directJSON](), "json", "unmarshal")
		if err != nil {
			return err
		}
		target := &directJSON{}
		if err := unmarshalTr.ReadAs(context.Background(), io.NopCloser(strings.NewReader(`{"value":"x"}`)), target); err == nil {
			return errTransformer("expected direct unmarshal error for unsupported target")
		}
		return nil
	}))

	Then(t, "json direct marshal 错误会继续上抛", ExpectMust(func() error {
		tr, err := contentinternal.New(reflect.TypeFor[failingJSON](), "json", "marshal")
		if err != nil {
			return err
		}
		if _, err := tr.Prepare(context.Background(), failingJSON{}); err == nil || !strings.Contains(err.Error(), "marshal failed") {
			return errTransformer("expected direct marshal failure")
		}
		return nil
	}))
}

func TestOctetAndMultipartBranches(t *testing.T) {
	t.Run("octet read sets filename and content type", func(t *testing.T) {
		Then(t, "octet transformer 会透传 header 元数据给目标对象", ExpectMust(func() error {
			tr, err := contentinternal.New(reflect.TypeFor[octetMetaTarget](), "octet", "unmarshal")
			if err != nil {
				return err
			}

			header := http.Header{}
			header.Set("Content-Type", "text/plain")
			header.Set("Content-Disposition", `attachment; filename="demo.txt"`)

			target := &octetMetaTarget{}
			if err := tr.ReadAs(context.Background(), contentinternal.ReadCloseWithHeader(io.NopCloser(strings.NewReader("payload")), header), target); err != nil {
				return err
			}
			if target.Name != "demo.txt" || target.Type != "text/plain" || string(target.Data) != "payload" {
				return errTransformer("unexpected octet metadata target")
			}
			return nil
		}))
	})

	t.Run("octet read supports direct passthrough targets", func(t *testing.T) {
		Then(t, "octet transformer 可直接写入 io.Writer 或透传 io.ReadCloser", ExpectMust(func() error {
			tr, err := contentinternal.New(reflect.TypeFor[[]byte](), "octet", "unmarshal")
			if err != nil {
				return err
			}

			buf := bytes.NewBuffer(nil)
			if err := tr.ReadAs(context.Background(), io.NopCloser(strings.NewReader("writer")), buf); err != nil {
				return err
			}
			if buf.String() != "writer" {
				return errTransformer("unexpected octet writer result")
			}

			var rc io.ReadCloser
			if err := tr.ReadAs(context.Background(), io.NopCloser(strings.NewReader("closer")), &rc); err != nil {
				return err
			}
			defer rc.Close()

			data, err := io.ReadAll(rc)
			if err != nil {
				return err
			}
			if string(data) != "closer" {
				return errTransformer("unexpected octet read closer result")
			}
			return nil
		}))
	})

	t.Run("multipart read rejects non pointer target", func(t *testing.T) {
		Then(t, "multipart transformer 要求 pointer 目标值", ExpectMust(func() error {
			tr, err := contentinternal.New(reflect.TypeFor[struct{ Name string }](), "multipart", "unmarshal")
			if err != nil {
				return err
			}

			buf := bytes.NewBuffer(nil)
			mw := multipart.NewWriter(buf)
			_ = mw.Close()

			header := http.Header{}
			header.Set("Content-Type", mw.FormDataContentType())

			if err := tr.ReadAs(context.Background(), contentinternal.ReadCloseWithHeader(io.NopCloser(buf), header), struct{ Name string }{}); err == nil || !strings.Contains(err.Error(), "ptr value") {
				return errTransformer("expected non-pointer multipart error")
			}
			return nil
		}))
	})
}

func TestTextOctetAndURLEncodedRuntimeBranches(t *testing.T) {
	t.Run("text transformer supports bytes read and struct marshal", func(t *testing.T) {
		Then(t, "text transformer 可读入 []byte，并把结构体按文本值输出", ExpectMust(func() error {
			readTr, err := contentinternal.New(reflect.TypeFor[[]byte](), "text", "unmarshal")
			if err != nil {
				return err
			}

			var target []byte
			if err := readTr.ReadAs(context.Background(), io.NopCloser(strings.NewReader("hello")), &target); err != nil {
				return err
			}
			if string(target) != "hello" {
				return errTransformer("unexpected text read result")
			}

			writeTr, err := contentinternal.New(reflect.TypeFor[stringAlias](), "text", "marshal")
			if err != nil {
				return err
			}
			content, err := writeTr.Prepare(context.Background(), stringAlias("demo"))
			if err != nil {
				return err
			}
			defer content.Close()

			data, err := io.ReadAll(content)
			if err != nil {
				return err
			}
			if string(data) != "demo" {
				return errTransformer("unexpected text marshal result")
			}
			return nil
		}))
	})

	t.Run("octet prepare supports nil reader and structured fallback", func(t *testing.T) {
		Then(t, "octet transformer 可处理 nil、reader 和默认 JSON 反引号分支", ExpectMust(func() error {
			nilTr, err := contentinternal.New(reflect.TypeFor[[]byte](), "octet", "marshal")
			if err != nil {
				return err
			}
			nilContent, err := nilTr.Prepare(context.Background(), nil)
			if err != nil {
				return err
			}
			defer nilContent.Close()
			if nilContent.GetContentLength() != 0 {
				return errTransformer("unexpected nil octet content length")
			}

			readerContent, err := nilTr.Prepare(context.Background(), strings.NewReader("stream"))
			if err != nil {
				return err
			}
			defer readerContent.Close()
			readerData, err := io.ReadAll(readerContent)
			if err != nil {
				return err
			}
			if string(readerData) != "stream" {
				return errTransformer("unexpected octet reader content")
			}

			structContent, err := nilTr.Prepare(context.Background(), stringAlias("demo"))
			if err != nil {
				return err
			}
			defer structContent.Close()
			structData, err := io.ReadAll(structContent)
			if err != nil {
				return err
			}
			if string(structData) != "demo" {
				return errTransformer("unexpected octet fallback content")
			}
			return nil
		}))
	})

	t.Run("urlencoded transformer supports struct round trip", func(t *testing.T) {
		type payload struct {
			Name string `json:"name"`
			Age  int    `json:"age"`
		}

		Then(t, "urlencoded transformer 可在 struct 与 query-string 之间转换", ExpectMust(func() error {
			readTr, err := contentinternal.New(reflect.TypeFor[payload](), "urlencoded", "unmarshal")
			if err != nil {
				return err
			}

			var decoded payload
			if err := readTr.ReadAs(context.Background(), io.NopCloser(strings.NewReader("name=demo&age=2")), &decoded); err != nil {
				return err
			}
			if decoded.Name != "demo" || decoded.Age != 2 {
				return errTransformer("unexpected urlencoded decode result")
			}

			writeTr, err := contentinternal.New(reflect.TypeFor[payload](), "urlencoded", "marshal")
			if err != nil {
				return err
			}
			content, err := writeTr.Prepare(context.Background(), payload{Name: "demo", Age: 2})
			if err != nil {
				return err
			}
			defer content.Close()
			data, err := io.ReadAll(content)
			if err != nil {
				return err
			}

			values, err := url.ParseQuery(string(data))
			if err != nil {
				return err
			}
			if values.Get("name") != "demo" || values.Get("age") != "2" {
				return errTransformer("unexpected urlencoded encode result")
			}
			return nil
		}))
	})
}

func errTransformer(msg string) error {
	return &transformerErr{msg: msg}
}

type transformerErr struct{ msg string }

func (e *transformerErr) Error() string { return e.msg }
