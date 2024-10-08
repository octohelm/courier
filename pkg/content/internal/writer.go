package internal

import (
	"io"
	"sync"
)

func DeferWriter(createWriter func() (io.Writer, error)) io.Writer {
	return &deferWriter{
		createWriter: createWriter,
	}
}

type deferWriter struct {
	createWriter func() (io.Writer, error)

	once sync.Once
	w    io.Writer
	err  error
}

func (d *deferWriter) Write(p []byte) (n int, err error) {
	d.once.Do(func() {
		d.w, d.err = d.createWriter()
	})
	if d.err != nil {
		return -1, d.err
	}
	return d.w.Write(p)
}
