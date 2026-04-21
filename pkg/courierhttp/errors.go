package courierhttp

import "fmt"

// ErrContextCanceled 表示上下文取消错误。
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
