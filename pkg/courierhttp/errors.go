package courierhttp

import "fmt"

type ErrContextCanceled struct {
	Reason string
}

func (e *ErrContextCanceled) StatusCode() int {
	// https://httpstatuses.com/499
	return 499
}

func (e *ErrContextCanceled) Error() string {
	return fmt.Sprintf("context canceled: %s", e.Reason)
}
