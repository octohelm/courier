package testingutil

import (
	"os"

	"github.com/go-json-experiment/json"
)

func MustJSONRaw(v interface{}) string {
	data, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return string(data)
}

func PrintJSON(v interface{}) {
	// FIXME set option until https://github.com/go-json-experiment/json/pull/20
	_ = json.MarshalWrite(os.Stdout, v)
}
