package testingutil

import (
	"bytes"
	"fmt"
	"os"

	"github.com/go-json-experiment/json"
)

func PrintJSON(v any) {
	if err := json.MarshalWrite(os.Stdout, v); err != nil {
		panic(err)
	}
}

func BeJSON[X any](expect string) func(v X) error {
	expectData := bytes.TrimSpace([]byte(expect))

	return func(v X) error {
		actual, err := json.Marshal(v)
		if err != nil {
			return err
		}

		if bytes.Equal(actual, expectData) {
			return nil
		}

		return fmt.Errorf("json mismatch\nexpect:\n%s\nactual:\n%s", expectData, actual)
	}
}
