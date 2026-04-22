package internal_test

import (
	"reflect"
	"strings"
	"testing"

	"github.com/octohelm/courier/pkg/content/internal"
)

import (
	_ "github.com/octohelm/courier/pkg/content/transformers"
)

type textAlias string

func (m textAlias) MarshalText() ([]byte, error) {
	return []byte(string(m)), nil
}

type textAliasInput string

func (m *textAliasInput) UnmarshalText(text []byte) error {
	*m = textAliasInput(text)
	return nil
}

type jsonAlias struct {
	Value string `json:"value"`
}

func (v *jsonAlias) UnmarshalJSON(data []byte) error {
	v.Value = string(data)
	return nil
}

func TestTransformerSelection(t *testing.T) {
	cases := []struct {
		name      string
		typ       reflect.Type
		mediaType string
		action    string
		expect    string
	}{
		{name: "string marshal", typ: reflect.TypeFor[string](), action: "marshal", expect: "text/plain"},
		{name: "bytes marshal", typ: reflect.TypeFor[[]byte](), action: "marshal", expect: "application/octet-stream"},
		{name: "text marshal", typ: reflect.TypeFor[textAlias](), action: "marshal", expect: "text/plain"},
		{name: "text unmarshal", typ: reflect.TypeFor[textAliasInput](), action: "unmarshal", expect: "text/plain"},
		{name: "json unmarshal", typ: reflect.TypeFor[jsonAlias](), action: "unmarshal", expect: "application/json"},
		{name: "+json alias", typ: reflect.TypeFor[struct{ Name string }](), mediaType: "application/merge-patch+json", action: "marshal", expect: "application/json"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tr, err := internal.New(tc.typ, tc.mediaType, tc.action)
			if err != nil {
				t.Fatalf("unexpected new transformer error: %v", err)
			}
			if tr.MediaType() != tc.expect {
				t.Fatalf("unexpected media type: %s", tr.MediaType())
			}
		})
	}
}

func TestTransformerSelectionReturnsErrorForUnknownMediaType(t *testing.T) {
	_, err := internal.New(reflect.TypeFor[string](), "application/unknown", "marshal")
	if err == nil || !strings.Contains(err.Error(), "unknown media type") {
		t.Fatalf("unexpected error: %v", err)
	}
}
