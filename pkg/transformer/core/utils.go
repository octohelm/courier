package core

import (
	"io"
	"net/http"
	"net/textproto"
)

func MIMEHeader(headers ...textproto.MIMEHeader) textproto.MIMEHeader {
	header := textproto.MIMEHeader{}
	for _, h := range headers {
		for k, values := range h {
			for _, v := range values {
				header.Add(k, v)
			}
		}
	}
	return header
}

type WithHeader interface {
	Header() http.Header
}

type HeaderWriter interface {
	WithHeader
	io.Writer
}

func WriterWithHeader(w io.Writer, header http.Header) HeaderWriter {
	return &headerWriter{Writer: w, header: header}
}

func (f *headerWriter) Header() http.Header {
	return f.header
}

type headerWriter struct {
	io.Writer
	header http.Header
}
