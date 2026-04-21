package courierhttp

import (
	"net/url"
)

// Redirect 创建重定向响应。
func Redirect(statusCode int, location *url.URL) Response[any] {
	return &response[any]{
		statusCode: statusCode,
		location:   location,
	}
}
