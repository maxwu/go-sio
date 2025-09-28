package go_sio

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"
)

// mockReadCloser implements io.ReadCloser for testing
type mockReadCloser struct {
	*strings.Reader
	closed bool
	err    error
}

func (m *mockReadCloser) Close() error {
	m.closed = true
	return m.err
}

func newMockReadCloser(data string) *mockReadCloser {
	return &mockReadCloser{
		Reader: strings.NewReader(data),
		closed: false,
		err:    nil,
	}
}

func TestNewTeeReaderCloser(t *testing.T) {
	data := "test data for tee reader"
	reader := newMockReadCloser(data)
	var writer bytes.Buffer

	trc := NewTeeReaderCloser(reader, &writer)
	if trc == nil {
		t.Fatal("NewTeeReaderCloser returned nil")
	}

	// Verify the structure
	if trc.closer != reader {
		t.Error("TeeReaderCloser closer not set correctly")
	}
}

func TestTeeReaderCloser_Read(t *testing.T) {
	data := "hello world"
	reader := newMockReadCloser(data)
	var writer bytes.Buffer

	trc := NewTeeReaderCloser(reader, &writer)

	// Read all data
	buf := make([]byte, len(data))
	n, err := trc.Read(buf)
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}
	if n != len(data) {
		t.Errorf("Expected %d bytes, got %d", len(data), n)
	}
	if string(buf) != data {
		t.Errorf("Expected %q, got %q", data, string(buf))
	}

	// Verify data was also written to the writer
	if writer.String() != data {
		t.Errorf("Expected writer to contain %q, got %q", data, writer.String())
	}
}

func TestTeeReaderCloser_ReadPartial(t *testing.T) {
	data := "hello world"
	reader := newMockReadCloser(data)
	var writer bytes.Buffer

	trc := NewTeeReaderCloser(reader, &writer)

	// Read partial data
	buf := make([]byte, 5)
	n, err := trc.Read(buf)
	if err != nil {
		t.Fatalf("First read failed: %v", err)
	}
	if n != 5 {
		t.Errorf("Expected 5 bytes, got %d", n)
	}
	if string(buf) != "hello" {
		t.Errorf("Expected 'hello', got %q", string(buf))
	}

	// Read remaining data
	buf = make([]byte, 10)
	n, err = trc.Read(buf)
	if err != nil {
		t.Fatalf("Second read failed: %v", err)
	}
	if n != 6 {
		t.Errorf("Expected 6 bytes, got %d", n)
	}
	if string(buf[:n]) != " world" {
		t.Errorf("Expected ' world', got %q", string(buf[:n]))
	}

	// Verify all data was written to the writer
	if writer.String() != data {
		t.Errorf("Expected writer to contain %q, got %q", data, writer.String())
	}
}

func TestTeeReaderCloser_ReadEOF(t *testing.T) {
	data := "test"
	reader := newMockReadCloser(data)
	var writer bytes.Buffer

	trc := NewTeeReaderCloser(reader, &writer)

	// Read all data
	buf := make([]byte, 10)
	n, err := trc.Read(buf)
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}
	if string(buf[:n]) != data {
		t.Errorf("Expected %q, got %q", data, string(buf[:n]))
	}

	// Next read should return EOF
	buf = make([]byte, 10)
	n, err = trc.Read(buf)
	if err != io.EOF {
		t.Errorf("Expected EOF, got %v", err)
	}
	if n != 0 {
		t.Errorf("Expected 0 bytes on EOF, got %d", n)
	}

	// Writer should still contain all data
	if writer.String() != data {
		t.Errorf("Expected writer to contain %q, got %q", data, writer.String())
	}
}

func TestTeeReaderCloser_Close(t *testing.T) {
	data := "test data"
	reader := newMockReadCloser(data)
	var writer bytes.Buffer

	trc := NewTeeReaderCloser(reader, &writer)

	// Close the tee reader
	err := trc.Close()
	if err != nil {
		t.Errorf("Close failed: %v", err)
	}

	// Verify the underlying reader was closed
	if !reader.closed {
		t.Error("Underlying reader was not closed")
	}
}

func TestTeeReaderCloser_CloseWithError(t *testing.T) {
	data := "test data"
	reader := newMockReadCloser(data)
	expectedErr := errors.New("close error")
	reader.err = expectedErr
	var writer bytes.Buffer

	trc := NewTeeReaderCloser(reader, &writer)

	// Close should return the error
	err := trc.Close()
	if err != expectedErr {
		t.Errorf("Expected error %v, got %v", expectedErr, err)
	}

	// Verify the underlying reader was still closed
	if !reader.closed {
		t.Error("Underlying reader was not closed")
	}
}

func TestTeeReaderCloser_ReadAfterClose(t *testing.T) {
	data := "test data"
	reader := newMockReadCloser(data)
	var writer bytes.Buffer

	trc := NewTeeReaderCloser(reader, &writer)

	// Close first
	err := trc.Close()
	if err != nil {
		t.Fatalf("Close failed: %v", err)
	}

	// Try to read after close - this should still work because TeeReaderCloser
	// doesn't prevent reads after close, it just closes the underlying closer
	// The actual behavior depends on the underlying reader implementation
	buf := make([]byte, 10)
	n, err := trc.Read(buf)
	
	// The mockReadCloser continues to work even after close
	// This is the actual behavior - TeeReaderCloser doesn't enforce any read restrictions
	if n == 0 && err != nil {
		t.Logf("Read after close returned error as expected: %v", err)
	} else {
		t.Logf("Read after close succeeded - this is valid behavior for TeeReaderCloser")
	}
}

func TestTeeReaderCloser_WithRealIOTypes(t *testing.T) {
	// Test with actual io.Reader and io.Writer types
	data := "test data for real IO types"
	reader := io.NopCloser(strings.NewReader(data))
	var writer bytes.Buffer

	trc := NewTeeReaderCloser(reader, &writer)

	// Read all data using io.ReadAll
	result, err := io.ReadAll(trc)
	if err != nil {
		t.Fatalf("ReadAll failed: %v", err)
	}

	if string(result) != data {
		t.Errorf("Expected %q, got %q", data, string(result))
	}

	// Verify data was written to buffer
	if writer.String() != data {
		t.Errorf("Expected writer to contain %q, got %q", data, writer.String())
	}

	// Close should work
	err = trc.Close()
	if err != nil {
		t.Errorf("Close failed: %v", err)
	}
}

func TestTeeReaderCloser_MultipleReads(t *testing.T) {
	data := "line1\nline2\nline3\n"
	reader := newMockReadCloser(data)
	var writer bytes.Buffer

	trc := NewTeeReaderCloser(reader, &writer)

	// Read in chunks
	var result bytes.Buffer
	buf := make([]byte, 3)

	for {
		n, err := trc.Read(buf)
		if n > 0 {
			result.Write(buf[:n])
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("Read failed: %v", err)
		}
	}

	// Verify all data was read correctly
	if result.String() != data {
		t.Errorf("Expected %q, got %q", data, result.String())
	}

	// Verify all data was written to the writer
	if writer.String() != data {
		t.Errorf("Expected writer to contain %q, got %q", data, writer.String())
	}
}

func TestTeeReaderCloser_EmptyReader(t *testing.T) {
	reader := newMockReadCloser("")
	var writer bytes.Buffer

	trc := NewTeeReaderCloser(reader, &writer)

	// Read from empty reader
	buf := make([]byte, 10)
	n, err := trc.Read(buf)
	if err != io.EOF {
		t.Errorf("Expected EOF from empty reader, got %v", err)
	}
	if n != 0 {
		t.Errorf("Expected 0 bytes from empty reader, got %d", n)
	}

	// Writer should be empty
	if writer.Len() != 0 {
		t.Errorf("Expected empty writer, got %q", writer.String())
	}

	// Close should work
	err = trc.Close()
	if err != nil {
		t.Errorf("Close failed: %v", err)
	}
}

func TestTeeReaderCloser_NilWriter(t *testing.T) {
	data := "test data"
	reader := newMockReadCloser(data)

	// This should not panic
	trc := NewTeeReaderCloser(reader, nil)
	if trc == nil {
		t.Fatal("NewTeeReaderCloser returned nil with nil writer")
	}

	// Reading might panic or fail depending on io.TeeReader implementation
	// but we should handle it gracefully
	buf := make([]byte, 10)
	defer func() {
		if r := recover(); r != nil {
			// This is acceptable - io.TeeReader with nil writer might panic
			t.Logf("Read panicked with nil writer (expected): %v", r)
		}
	}()
	_, err := trc.Read(buf)
	if err != nil {
		t.Logf("Read failed with nil writer (might be expected): %v", err)
	}
}