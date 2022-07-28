package testingutil

import (
	"encoding/json"
	"os"
)

func MustJSONRaw(v interface{}) string {
	data, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return string(data)
}

func PrintJSON(v interface{}) {
	e := json.NewEncoder(os.Stdout)
	e.SetIndent("", "  ")

	_ = e.Encode(v)
}
