package httprequest

import (
	"io"
	"net/http"
)

func WithHeader(rc io.ReadCloser, header http.Header) io.ReadCloser {
	return &readerWithHeader{
		ReadCloser: rc,
		header:     header,
	}
}

type readerWithHeader struct {
	header http.Header

	io.ReadCloser
}

func (b *readerWithHeader) Header() http.Header {
	return b.header
}
