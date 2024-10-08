package internal

import (
	"io"

	"golang.org/x/sync/errgroup"
)

type ReadCloserFrom interface {
	ReadFromCloser(r io.ReadCloser) (n int64, err error)
}

func Pipe(w func(w io.Writer) error, r func(r io.Reader) error) error {
	pr, pw := io.Pipe()
	wg := &errgroup.Group{}

	wg.Go(func() error {
		defer pr.Close()

		return r(pr)
	})

	wg.Go(func() error {
		defer pw.Close()

		return w(pw)
	})

	return wg.Wait()
}
