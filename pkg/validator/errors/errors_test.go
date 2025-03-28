package errors

import (
	"fmt"
)

func ExampleErrMissingRequired() {
	fmt.Println(&ErrMissingRequired{})
	// Output:
	// missing required field
}

func ExampleErrPatternNotMatch() {
	fmt.Println(&ErrPatternNotMatch{
		Subject: "value",
		Target:  "1",
		Pattern: `/\d+/`,
	})
	// Output:
	// value should match /\d+/, but got 1
}

func ExampleErrMultipleOf() {
	fmt.Println(&ErrMultipleOf{
		Subject:    "value",
		Target:     "11",
		MultipleOf: 2,
	})
	// Output:
	// value should be multiple of 2, but got 11
}

func ExampleErrNotInEnum() {
	fmt.Println(&ErrNotInEnum{
		Subject: "value",
		Target:  "11",
		Enums: []any{
			"1", "2", "3",
		},
	})
	// Output:
	// value should be one of 1, 2, 3, but got 11
}

func ExampleErrOutOfRange() {
	fmt.Println(&ErrOutOfRange{
		Subject:          "value",
		Minimum:          "1",
		Maximum:          "10",
		Target:           "11",
		ExclusiveMinimum: true,
		ExclusiveMaximum: true,
	})
	// Output:
	// value should be larger than 1 and less than 10, but got 11
}
