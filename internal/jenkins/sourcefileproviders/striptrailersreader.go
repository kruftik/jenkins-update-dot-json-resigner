package sourcefileproviders

import (
	"fmt"
	"io"
)

var (
	WrappedJSONPPrefix = []byte("updateCenter.post(\n")
	WrappedJSONPSuffix = []byte("\n);")

	WrappedHTMLPrefix = []byte("<!DOCTYPE html><html><head><meta http-equiv='Content-Type' content='text/html;charset=UTF-8' /></head><body><script>window.onload = function () { window.parent.postMessage(JSON.stringify(\n")
	WrappedHTMLSuffix = []byte("\n),'*'); };</script></body></html>")

	_ io.ReadCloser = StripTrailersReader{}
)

type reader interface {
	io.Closer
	io.ReaderAt
}

type readerOuter interface {
	io.Reader
	Outer() (io.ReaderAt, int64, int64)
}

type StripTrailersReader struct {
	r readerOuter
}

func NewJSONPTrailersStrippingReader(r reader, length int64) (io.ReadCloser, error) {
	str := StripTrailersReader{
		r: io.NewSectionReader(r, int64(len(WrappedJSONPPrefix)), length-int64(len(WrappedJSONPPrefix))-int64(len(WrappedJSONPSuffix))),
	}

	return str, nil
}

func (s StripTrailersReader) Read(p []byte) (n int, err error) {
	return s.r.Read(p)
}

func (s StripTrailersReader) Close() error {
	r, _, _ := s.r.Outer()

	closer, ok := r.(io.Closer)
	if !ok {
		return fmt.Errorf("cannot close outer reader")
	}

	if err := closer.Close(); err != nil {
		return fmt.Errorf("cannot close outer reader: %w", err)
	}

	return nil
}
