package internal

import (
	"io"

	"net/http"
)

func ReadCloseWithHeader(rc io.ReadCloser, h http.Header) interface {
	io.ReadCloser
	HeaderGetter
} {
	return &readCloserWithHeader{
		h:          h,
		ReadCloser: rc,
	}
}

type readCloserWithHeader struct {
	h http.Header
	io.ReadCloser
}

func (r *readCloserWithHeader) Header() http.Header {
	return r.h
}
