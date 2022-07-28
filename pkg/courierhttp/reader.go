package courierhttp

import (
	"io"
	"io/fs"
)

func WrapReadCloser(r io.Reader) io.ReadCloser {
	return &reader{Reader: r}
}

type reader struct {
	io.Reader
}

func (r *reader) Len() int64 {
	switch x := r.Reader.(type) {
	case interface{ Len() int64 }:
		return x.Len()
	case interface{ Len() int }:
		return int64(x.Len())
	case interface{ Stat() (fs.FileInfo, error) }:
		info, _ := x.Stat()
		return info.Size()
	}
	return 0
}

func (r *reader) Close() error {
	if c, ok := r.Reader.(io.ReadCloser); ok {
		return c.Close()
	}
	return nil
}
