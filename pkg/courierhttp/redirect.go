package courierhttp

import (
	"net/url"
)

func Redirect(statusCode int, location *url.URL) Response[any] {
	return &response[any]{
		statusCode: statusCode,
		location:   location,
	}
}
