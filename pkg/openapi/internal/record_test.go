package internal

import (
	"bytes"
	"testing"

	"github.com/go-json-experiment/json"
	"github.com/go-json-experiment/json/jsontext"
)

func TestRecordCRUDAndJSON(t *testing.T) {
	t.Run("supports ordered set get delete and key iteration", func(t *testing.T) {
		record := &Record[string, int]{}

		if !record.IsZero() {
			t.Fatalf("expected zero record")
		}
		if added := record.Set("a", 1); !added {
			t.Fatalf("expected first set to add key")
		}
		if added := record.Set("b", 2); !added {
			t.Fatalf("expected second set to add key")
		}
		if added := record.Set("a", 3); added {
			t.Fatalf("expected duplicate set to update existing key")
		}
		if record.Len() != 2 {
			t.Fatalf("unexpected record length: %d", record.Len())
		}
		if record.IsZero() {
			t.Fatalf("expected non-zero record")
		}
		if value, ok := record.Get("a"); !ok || value != 3 {
			t.Fatalf("unexpected get result: value=%d ok=%v", value, ok)
		}
		if _, ok := record.Get("missing"); ok {
			t.Fatalf("expected missing key lookup to fail")
		}

		keys := make([]string, 0, 2)
		values := make([]int, 0, 2)
		for key, value := range record.KeyValues() {
			keys = append(keys, key)
			values = append(values, value)
		}
		if !equalStrings(keys, []string{"a", "b"}) {
			t.Fatalf("unexpected key order: %#v", keys)
		}
		if !equalInts(values, []int{3, 2}) {
			t.Fatalf("unexpected value order: %#v", values)
		}

		if !record.Delete("a") {
			t.Fatalf("expected delete existing key")
		}
		if record.Delete("missing") {
			t.Fatalf("expected delete missing key to be false")
		}
	})

	t.Run("marshals and unmarshals object payload", func(t *testing.T) {
		record := &Record[string, int]{}
		record.Set("first", 1)
		record.Set("second", 2)

		data, err := json.Marshal(record)
		if err != nil {
			t.Fatalf("unexpected marshal error: %v", err)
		}
		if string(data) != `{"first":1,"second":2}` {
			t.Fatalf("unexpected marshal result: %s", data)
		}

		var decoded Record[string, int]
		if err := json.Unmarshal(data, &decoded); err != nil {
			t.Fatalf("unexpected unmarshal error: %v", err)
		}
		if decoded.Len() != 2 {
			t.Fatalf("unexpected decoded length: %d", decoded.Len())
		}
		if value, ok := decoded.Get("second"); !ok || value != 2 {
			t.Fatalf("unexpected decoded value: value=%d ok=%v", value, ok)
		}
	})

	t.Run("rejects non-object json token", func(t *testing.T) {
		var decoded Record[string, int]
		err := decoded.UnmarshalJSONFrom(jsontext.NewDecoder(bytes.NewBufferString(`[]`)))
		if err == nil {
			t.Fatalf("expected semantic error")
		}
	})

	t.Run("accepts empty payload as no-op", func(t *testing.T) {
		var decoded Record[string, int]
		if err := decoded.UnmarshalJSONFrom(jsontext.NewDecoder(bytes.NewBuffer(nil))); err != nil {
			t.Fatalf("unexpected empty input error: %v", err)
		}
	})

	t.Run("returns marshal error from unsupported key or value", func(t *testing.T) {
		valueRecord := Record[string, chan int]{}
		valueRecord.Set("broken", make(chan int))
		if err := valueRecord.MarshalJSONTo(jsontext.NewEncoder(bytes.NewBuffer(nil))); err == nil {
			t.Fatalf("expected marshal error for unsupported value")
		}

		keyRecord := Record[chan int, string]{}
		keyRecord.Set(make(chan int), "broken")
		if err := keyRecord.MarshalJSONTo(jsontext.NewEncoder(bytes.NewBuffer(nil))); err == nil {
			t.Fatalf("expected marshal error for unsupported key")
		}
	})
}

func equalStrings(a []string, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func equalInts(a []int, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
