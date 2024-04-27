package courierhttp

import "fmt"

type ErrContextCanceled struct {
	Reason string
}

func (ErrContextCanceled) StatusCode() int {
	// https://httpstatuses.com/499
	return 499
}

func (c *ErrContextCanceled) Error() string {
	return fmt.Sprintf("context canceled: %s", c.Reason)
}
