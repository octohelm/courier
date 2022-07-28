package core

import (
	"io"
	"strings"

	"github.com/pkg/errors"
)

type CanInterface interface {
	Interface() any
}

type CanString interface {
	String() string
}

func NewStringReaders(values []string) *StringReaders {
	bs := make([]io.ReadCloser, len(values))
	for i := range values {
		bs[i] = &StringReader{v: values[i]}
	}

	return &StringReaders{
		Readers: bs,
		values:  values,
	}
}

type StringReaders struct {
	idx     int
	Readers []io.ReadCloser
	values  []string
}

func (v *StringReaders) Close() error {
	for i := range v.Readers {
		_ = v.Readers[i].Close()
	}
	return nil
}

func (v *StringReaders) Interface() any {
	return v.values
}

func (v *StringReaders) Len() int {
	return len(v.Readers)
}

func (v *StringReaders) Read(p []byte) (n int, err error) {
	if v.idx < len(v.Readers) {
		return v.Readers[v.idx].Read(p)
	}
	return -1, errors.Errorf("bounds out of range, %d", v.idx)
}

func (v *StringReaders) NextReader() io.ReadCloser {
	r := v.Readers[v.idx]
	v.idx++
	return r
}

func NewStringReader(v string) *StringReader {
	return &StringReader{v: v}
}

type StringReader struct {
	v string
	r io.Reader
}

func (r *StringReader) Read(p []byte) (n int, err error) {
	if r.r == nil {
		r.r = strings.NewReader(r.v)
	}
	return r.r.Read(p)
}

func (r *StringReader) Interface() any {
	return r.v
}

func (r *StringReader) String() string {
	return r.v
}

func (r *StringReader) Close() error {
	return nil
}

func NewStringBuilders() *StringBuilders {
	return &StringBuilders{}
}

type StringBuilders struct {
	idx     int
	buffers []*strings.Builder
}

func (v *StringBuilders) SetN(n int) {
	v.buffers = make([]*strings.Builder, n)
	v.idx = 0
	for i := range v.buffers {
		v.buffers[i] = &strings.Builder{}
	}
}
func (v *StringBuilders) NextWriter() io.Writer {
	if v.idx == 0 && len(v.buffers) == 0 {
		v.SetN(1)
	}
	r := v.buffers[v.idx]
	v.idx++
	return r
}

func (v *StringBuilders) Write(p []byte) (n int, err error) {
	if v.idx == 0 && len(v.buffers) == 0 {
		v.SetN(1)
	}
	return v.buffers[v.idx].Write(p)
}

func (v *StringBuilders) StringSlice() []string {
	values := make([]string, len(v.buffers))
	for i, b := range v.buffers {
		values[i] = b.String()
	}
	return values
}
