package transformers

import (
	"context"
	"io"

	"github.com/go-courier/logr"
)

func NewContent(contentType string) *Content {
	return &Content{
		contentType:   contentType,
		contentLength: -1,
	}
}

type Content struct {
	contentType   string
	contentLength int64
	io.ReadCloser
}

func (c Content) GetContentType() string {
	return c.contentType
}

func (c Content) GetContentLength() int64 {
	return c.contentLength
}

func (c *Content) SetContentLength(n int64) {
	c.contentLength = n
}

func AsReaderCloser(ctx context.Context, createWriter func(w io.WriteCloser) func() error) io.ReadCloser {
	pr, pw := io.Pipe()

	write := createWriter(pw)

	go func() {
		if err := write(); err != nil {
			logr.FromContext(ctx).Error(err)
		}
	}()

	return pr
}
