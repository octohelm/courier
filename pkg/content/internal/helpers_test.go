package internal

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"reflect"
	"strings"
	"testing"

	. "github.com/octohelm/x/testing/v2"

	"github.com/octohelm/courier/internal/httprequest"
	"github.com/octohelm/courier/internal/jsonflags"
)

func TestReadCloseWithHeader(t *testing.T) {
	Then(t, "ReadCloseWithHeader 会暴露 header 信息", ExpectMust(func() error {
		h := http.Header{"Content-Type": []string{"text/plain"}}
		rc := ReadCloseWithHeader(io.NopCloser(strings.NewReader("demo")), h)

		withHeader, ok := rc.(HeaderGetter)
		if !ok {
			return errContentInternal("missing HeaderGetter")
		}
		if withHeader.Header().Get("Content-Type") != "text/plain" {
			return errContentInternal("unexpected header value")
		}
		data, err := io.ReadAll(rc)
		if err != nil {
			return err
		}
		if string(data) != "demo" {
			return errContentInternal("unexpected read content")
		}
		return nil
	}))
}

func TestDeferWriter(t *testing.T) {
	t.Run("creates writer lazily", func(t *testing.T) {
		created := 0
		buf := bytes.NewBuffer(nil)

		Then(t, "首次写入时才创建 writer，且只创建一次", ExpectMust(func() error {
			w := DeferWriter(func() (io.Writer, error) {
				created++
				return buf, nil
			})

			if created != 0 {
				return errContentInternal("writer should not be created eagerly")
			}
			if _, err := w.Write([]byte("a")); err != nil {
				return err
			}
			if _, err := w.Write([]byte("b")); err != nil {
				return err
			}
			if created != 1 || buf.String() != "ab" {
				return errContentInternal("unexpected lazy writer behavior")
			}
			return nil
		}))
	})

	t.Run("propagates create error", func(t *testing.T) {
		Then(t, "创建 writer 失败时返回错误", ExpectMust(func() error {
			w := DeferWriter(func() (io.Writer, error) {
				return nil, errors.New("create failed")
			})
			if _, err := w.Write([]byte("x")); err == nil || !strings.Contains(err.Error(), "create failed") {
				return errContentInternal("expected create error")
			}
			return nil
		}))
	})
}

func TestPipe(t *testing.T) {
	t.Run("streams between writer and reader", func(t *testing.T) {
		Then(t, "Pipe 会连接写端和读端", ExpectMust(func() error {
			return Pipe(
				func(w io.Writer) error {
					_, err := io.WriteString(w, "payload")
					return err
				},
				func(r io.Reader) error {
					data, err := io.ReadAll(r)
					if err != nil {
						return err
					}
					if string(data) != "payload" {
						return errContentInternal("unexpected pipe payload")
					}
					return nil
				},
			)
		}))
	})

	t.Run("returns writer error", func(t *testing.T) {
		Then(t, "写端错误会被返回", ExpectMust(func() error {
			err := Pipe(
				func(w io.Writer) error { return errors.New("writer failed") },
				func(r io.Reader) error {
					_, _ = io.ReadAll(r)
					return nil
				},
			)
			if err == nil || !strings.Contains(err.Error(), "writer failed") {
				return errContentInternal("expected writer failure")
			}
			return nil
		}))
	})
}

func TestRequestUnmarshalRequestInfo(t *testing.T) {
	type requestPayload struct {
		Body string `in:"body"`
	}

	t.Run("delegates from http.Request wrapper", func(t *testing.T) {
		Then(t, "UnmarshalRequest 会通过 httprequest.From 走统一逻辑", ExpectMust(func() error {
			req, err := http.NewRequest(http.MethodPost, "/", io.NopCloser(strings.NewReader("demo")))
			if err != nil {
				return err
			}
			req.Header.Set("Content-Type", "text/plain")

			v := &requestPayload{}
			p := &Request{ParamValue{Value: reflect.ValueOf(v).Elem()}}
			if err := p.UnmarshalRequest(req); err != nil {
				return err
			}
			if v.Body != "demo" {
				return errContentInternal("unexpected body value")
			}
			return nil
		}))
	})

	t.Run("ignores missing body", func(t *testing.T) {
		type queryOnlyRequest struct {
			Query string `name:"query,omitzero" in:"query"`
		}

		Then(t, "未声明 body 字段时可在无 body 请求下正常解析", ExpectMust(func() error {
			v := &queryOnlyRequest{}
			p := &Request{ParamValue{Value: reflect.ValueOf(v).Elem()}}
			req, err := http.NewRequest(http.MethodGet, "/?query=demo", nil)
			if err != nil {
				return err
			}
			if err := p.UnmarshalRequestInfo(httprequest.From(req)); err != nil {
				return err
			}
			if v.Query != "demo" {
				return errContentInternal("unexpected query value")
			}
			return nil
		}))
	})
}

func TestParamValueReflectionHelpers(t *testing.T) {
	type payload struct {
		Name  string   `json:"name"`
		Items []string `json:"items"`
	}

	fields, err := jsonflags.Structs.StructFields(reflect.TypeFor[payload]())
	if err != nil {
		t.Fatalf("unexpected struct field error: %v", err)
	}

	var nameField, itemsField *jsonflags.StructField
	for sf := range fields.StructField() {
		switch sf.FieldName {
		case "Name":
			nameField = sf
		case "Items":
			itemsField = sf
		}
	}

	if nameField == nil || itemsField == nil {
		t.Fatalf("expected both fields to be discovered")
	}

	value := &ParamValue{Value: reflect.ValueOf(&payload{
		Name:  "demo",
		Items: []string{"a", "b"},
	}).Elem()}

	t.Run("Values iterates scalar and slice fields", func(t *testing.T) {
		names := make([]string, 0, 1)
		for rv := range value.Values(nameField) {
			names = append(names, rv.String())
		}
		if len(names) != 1 || names[0] != "demo" {
			t.Fatalf("unexpected scalar values: %#v", names)
		}

		items := make([]string, 0, 2)
		for rv := range value.Values(itemsField) {
			items = append(items, rv.String())
		}
		if !reflect.DeepEqual(items, []string{"a", "b"}) {
			t.Fatalf("unexpected slice values: %#v", items)
		}
	})

	t.Run("AddrValues exposes addressable elements", func(t *testing.T) {
		ptrs := make([]string, 0, 2)
		for _, rv := range value.AddrValues(itemsField, 2) {
			ptrs = append(ptrs, rv.Elem().String())
		}
		if !reflect.DeepEqual(ptrs, []string{"a", "b"}) {
			t.Fatalf("unexpected slice addr values: %#v", ptrs)
		}

		single := make([]string, 0, 1)
		for _, rv := range value.AddrValues(nameField, 1) {
			single = append(single, rv.Elem().String())
		}
		if !reflect.DeepEqual(single, []string{"demo"}) {
			t.Fatalf("unexpected scalar addr values: %#v", single)
		}
	})
}

func errContentInternal(msg string) error {
	return &contentInternalErr{msg: msg}
}

type contentInternalErr struct{ msg string }

func (e *contentInternalErr) Error() string { return e.msg }
