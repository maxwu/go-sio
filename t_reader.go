package go_sio

import "io"

type TeeReaderCloser struct {
	reader io.Reader
	closer io.Closer
}

func (t *TeeReaderCloser) Read(p []byte) (n int, err error) {
	return t.reader.Read(p)
}

func (t *TeeReaderCloser) Close() error {
	return t.closer.Close()
}

func NewTeeReaderCloser(r io.ReadCloser, w io.Writer) *TeeReaderCloser {
	return &TeeReaderCloser{reader: io.TeeReader(r, w), closer: r}
}
