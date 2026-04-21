package courier

import (
	"context"
	"io"
)

// Response 表示可声明响应体类型的请求契约。
type Response[T any] interface {
	ResponseData() T
}

// DoWith 执行客户端调用并解包响应数据。
func DoWith[Data any, Op Response[Data]](ctx context.Context, c Client, req Op, metas ...Metadata) (Data, error) {
	switch any(req).(type) {
	case interface{ ResponseData() *NoContent }:
		resp := req.ResponseData()
		_, err := c.Do(ctx, req, metas...).Into(nil)
		return resp, err
	case interface{ ResponseData() io.ReadCloser }:
		var body io.ReadCloser
		_, err := c.Do(ctx, req, metas...).Into(&body)
		if err != nil {
			var zero Data
			return zero, err
		}
		return any(body).(Data), nil
	}

	resp := req.ResponseData()
	_, err := c.Do(ctx, req, metas...).Into(resp)
	return resp, err
}

type NoContent struct{}
