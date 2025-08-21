package validator

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/octohelm/courier/pkg/statuserror"
	testingx "github.com/octohelm/x/testing"
)

type Named string

func (Named) StructTagValidate() string {
	return "@string[2,]"
}

type CustomType int

func (*CustomType) UnmarshalText(b []byte) error {
	return fmt.Errorf("invalid CustomType: %v", string(b))
}

func TestUnmarshal(t *testing.T) {
	type SubPtrStruct struct {
		PtrInt   *int     `validate:"@int[1,]"`
		PtrFloat *float32 `validate:"@float[1,]"`
		PtrUint  *uint    `validate:"@uint[1,]"`
	}

	type SubStruct struct {
		Int   int     `validate:"@int[1,]"`
		Float float32 `validate:"@float[1,]"`
		Uint  uint    `validate:"@uint[1,]"`
	}

	type SomeStruct struct {
		JustRequired string
		CanEmpty     *string `validate:"@string[0,]?"`
		String       string  `validate:"@string[1,]"`
		CustomType   CustomType
		Named        Named
		PtrString    *string              `validate:"@string[3,]" default:"123"`
		Slice        []string             `validate:"@slice<@string[1,]>"`
		SliceStruct  []SubStruct          `validate:"@slice"`
		Map          map[string]string    `validate:"@map<@string[2,],@string[1,]>"`
		MapStruct    map[string]SubStruct `validate:"@map<@string[2,],>"`
		Struct       SubStruct
		SubStruct
		*SubPtrStruct
	}

	v := &SomeStruct{}

	err := Unmarshal([]byte(`
{
	"Slice": ["", ""],
	"SliceStruct": [{ "Int": 0 }],
    "CustomType": "custom",
	"Named": "1",
	"Map": {
		"1":  "",
		"11": "",
		"12": ""
    },
	"MapStruct": {
		"x.io/x": {}
	}
}
`), v)

	b := bytes.NewBuffer(nil)

	for e := range statuserror.All(err) {
		_, _ = fmt.Fprintf(io.MultiWriter(b, os.Stdout), "%s\n", e)
	}

	testingx.Expect(t, strings.TrimSpace(b.String()), testingx.Be(strings.TrimSpace(`
string value length should be larger or equal than 1, but got 0 at /Slice/0
string value length should be larger or equal than 1, but got 0 at /Slice/1
integer value should be larger or equal than 1 and less or equal than 2147483647, but got 0 at /SliceStruct/0/Int
missing required field at /SliceStruct/0/Float
missing required field at /SliceStruct/0/Uint
invalid CustomType: custom at /CustomType
string value length should be larger or equal than 2, but got 1 at /Named
string value length should be larger or equal than 2, but got 1 at /Map/1/
string value length should be larger or equal than 1, but got 0 at /Map/1
string value length should be larger or equal than 1, but got 0 at /Map/11
string value length should be larger or equal than 1, but got 0 at /Map/12
missing required field at /MapStruct/x.io~1x/Int
missing required field at /MapStruct/x.io~1x/Float
missing required field at /MapStruct/x.io~1x/Uint
missing required field at /JustRequired
missing required field at /String
missing required field at /PtrString
missing required field at /Struct
missing required field at /Int
missing required field at /Float
missing required field at /Uint
missing required field at /PtrInt
missing required field at /PtrFloat
missing required field at /PtrUint
`)))
}
