package remoteurl

import (
	"fmt"
	"io"
	"os"
)

var (
	_ io.ReaderAt = TempFileReadCloser{}
	_ io.Closer   = TempFileReadCloser{}
)

type reader interface {
	io.Reader
	io.ReaderAt
	io.Closer
	Name() string
}

type TempFileReadCloser struct {
	f reader
}

func (f TempFileReadCloser) Read(p []byte) (n int, err error) {
	return f.f.Read(p)
}

func (f TempFileReadCloser) ReadAt(p []byte, off int64) (n int, err error) {
	return f.f.ReadAt(p, off)
}

func (f TempFileReadCloser) Close() error {
	if err := f.f.Close(); err != nil {
		return fmt.Errorf("cannot close file: %w", err)
	}

	if err := os.Remove(f.f.Name()); err != nil {
		return fmt.Errorf("cannot remove temporary file: %w", err)
	}

	return nil
}
