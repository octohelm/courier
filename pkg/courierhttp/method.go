package courierhttp

import "net/http"

// Method HTTP方法类型标记。
type Method struct{}

// MethodGet 表示HTTP GET方法。
type MethodGet struct{}

func (MethodGet) Method() string {
	return http.MethodGet
}

// MethodHead 表示HTTP HEAD方法。
type MethodHead struct{}

func (MethodHead) Method() string {
	return http.MethodHead
}

// MethodPost 表示HTTP POST方法。
type MethodPost struct{}

func (MethodPost) Method() string {
	return http.MethodPost
}

// MethodPut 表示HTTP PUT方法。
type MethodPut struct{}

func (MethodPut) Method() string {
	return http.MethodPut
}

// MethodPatch 表示HTTP PATCH方法。
type MethodPatch struct{}

func (MethodPatch) Method() string {
	return http.MethodPatch
}

// MethodDelete 表示HTTP DELETE方法。
type MethodDelete struct{}

func (MethodDelete) Method() string {
	return http.MethodDelete
}

// MethodConnect 表示HTTP CONNECT方法。
type MethodConnect struct{}

func (MethodConnect) Method() string {
	return http.MethodConnect
}

// MethodOptions 表示HTTP OPTIONS方法。
type MethodOptions struct{}

func (MethodOptions) Method() string {
	return http.MethodOptions
}

// MethodTrace 表示HTTP TRACE方法。
type MethodTrace struct{}

func (MethodTrace) Method() string {
	return http.MethodTrace
}
