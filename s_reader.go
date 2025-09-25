package go_sio

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"io"
)

var (
	ErrNilReader = errors.New("reader is nil")
	NopFilter StringLineFilter = func(in string) (string, error) { return in, nil }
)

type StringLineFilter func(string) (string, error)

type StreamReader struct {
	scanner    *bufio.Scanner
	filter     StringLineFilter
	buffer     bytes.Buffer
	existsData bool
}

func NewStreamReader(r io.Reader, f StringLineFilter) *StreamReader {
	if r == nil {
		return nil
	}
	if f == nil {
		f = NopFilter
	}
	sr := &StreamReader{
		scanner:    bufio.NewScanner(r),
		existsData: true,
		filter:     f,
	}

	sr.scanner.Split(split)
	return sr
}

func (sr *StreamReader) Read(p []byte) (n int, err error) {
	if sr == nil {
		return 0, ErrNilReader
	}
	var lineBytes []byte
	var lineStr string
	var bufErr error

	for sr.existsData && bufErr == nil {
		if sr.existsData = sr.scanner.Scan(); !sr.existsData {
			break
		}

		lineBytes = sr.scanner.Bytes()
		lineStr, bufErr = sr.filter(string(lineBytes))
		if bufErr != nil {
			break
		}
		if lineStr != "" {
			_, _ = sr.buffer.Write([]byte(lineStr))
			break
		}
	}

	if !sr.existsData && bufErr == nil {
		bufErr = sr.scanner.Err()
	}

	if bufErr == nil {
		return sr.buffer.Read(p)
	}
	return 0, bufErr
}

func split(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	if i := bytes.IndexByte(data, '\n'); i >= 0 {
		return i + 1, data[0 : i+1], nil
	}
	if atEOF {
		return len(data), data, nil
	}

	return 0, nil, nil
}

func NewJSONFilterReadCloser(r io.ReadCloser) io.ReadCloser {
	return NewReadCloser(
		NewStreamReader(
			r,
			func(in string) (string, error) {
				if json.Valid([]byte(in)) {
					return in, nil
				}
				return "", nil
			},
		),
		r,
	)
}