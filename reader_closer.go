package go_sio

import "io"

type ReadCloser struct {
	io.Reader
	io.Closer
}

func NewReadCloser(r io.Reader, c io.Closer) *ReadCloser {
	return &ReadCloser{Reader: r, Closer: c}
}
