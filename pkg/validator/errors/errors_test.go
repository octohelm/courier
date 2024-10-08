package errors

import (
	"fmt"
)

func ExampleErrMissingRequiredField() {
	fmt.Println(&ErrMissingRequired{})
	// Output:
	// missing required field
}

func ExampleErrNotMatch() {
	fmt.Println(&ErrNotMatch{
		Topic:   "value",
		Current: "1",
		Pattern: `/\d+/`,
	})
	// Output:
	// value should match /\d+/, but got 1
}

func ExampleMultipleOfError() {
	fmt.Println(&ErrMultipleOf{
		Topic:      "value",
		Current:    "11",
		MultipleOf: 2,
	})
	// Output:
	// value should be multiple of 2, but got 11
}

func ExampleNotInEnumError() {
	fmt.Println(&NotInEnumError{
		Topic:   "value",
		Current: "11",
		Enums: []any{
			"1", "2", "3",
		},
	})
	// Output:
	// value should be one of 1, 2, 3, but got 11
}

func ExampleOutOfRangeError() {
	fmt.Println(&OutOfRangeError{
		Topic:            "value",
		Minimum:          "1",
		Maximum:          "10",
		Current:          "11",
		ExclusiveMinimum: true,
		ExclusiveMaximum: true,
	})
	// Output:
	// value should be larger than 1 and less than 10, but got 11
}
