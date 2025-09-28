package go_sio

import (
	"errors"
	"io"
	"strings"
	"testing"
)

// Mock reader for testing
type mockReader struct {
	data string
	pos  int
}

func (m *mockReader) Read(p []byte) (n int, err error) {
	if m.pos >= len(m.data) {
		return 0, io.EOF
	}
	n = copy(p, m.data[m.pos:])
	m.pos += n
	return n, nil
}

// Mock closer for testing
type mockCloser struct {
	closed bool
	err    error
}

func (m *mockCloser) Close() error {
	m.closed = true
	return m.err
}

func TestNewReadCloser(t *testing.T) {
	reader := &mockReader{data: "test data"}
	closer := &mockCloser{}

	rc := NewReadCloser(reader, closer)
	if rc == nil {
		t.Fatal("NewReadCloser returned nil")
	}

	// Test that the reader works
	buf := make([]byte, 100)
	n, err := rc.Read(buf)
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}
	if string(buf[:n]) != "test data" {
		t.Errorf("Expected 'test data', got %q", string(buf[:n]))
	}

	// Test that the closer works
	err = rc.Close()
	if err != nil {
		t.Fatalf("Close failed: %v", err)
	}
	if !closer.closed {
		t.Error("Closer was not called")
	}
}

func TestReadCloser_ReadBehavior(t *testing.T) {
	// Test with multiple reads
	reader := &mockReader{data: "hello world"}
	closer := &mockCloser{}
	rc := NewReadCloser(reader, closer)

	// First read
	buf1 := make([]byte, 5)
	n1, err := rc.Read(buf1)
	if err != nil {
		t.Fatalf("First read failed: %v", err)
	}
	if n1 != 5 || string(buf1) != "hello" {
		t.Errorf("First read: expected 'hello' (5 bytes), got %q (%d bytes)", string(buf1), n1)
	}

	// Second read
	buf2 := make([]byte, 6)
	n2, err := rc.Read(buf2)
	if err != nil {
		t.Fatalf("Second read failed: %v", err)
	}
	if n2 != 6 || string(buf2) != " world" {
		t.Errorf("Second read: expected ' world' (6 bytes), got %q (%d bytes)", string(buf2), n2)
	}

	// Third read should return EOF
	buf3 := make([]byte, 10)
	n3, err := rc.Read(buf3)
	if err != io.EOF {
		t.Errorf("Expected EOF, got: %v", err)
	}
	if n3 != 0 {
		t.Errorf("Expected 0 bytes on EOF, got %d", n3)
	}
}

func TestReadCloser_CloseWithError(t *testing.T) {
	expectedErr := errors.New("close error")
	reader := &mockReader{data: "test"}
	closer := &mockCloser{err: expectedErr}
	rc := NewReadCloser(reader, closer)

	err := rc.Close()
	if err != expectedErr {
		t.Errorf("Expected close error %v, got %v", expectedErr, err)
	}
	if !closer.closed {
		t.Error("Closer was not called")
	}
}

func TestReadCloser_WithNilComponents(t *testing.T) {
	// Test with nil reader - should panic when trying to read
	closer := &mockCloser{}
	rc := NewReadCloser(nil, closer)
	if rc == nil {
		t.Fatal("NewReadCloser returned nil with nil reader")
	}

	buf := make([]byte, 10)
	
	// Reading from nil reader should panic
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic when reading from nil reader")
		}
	}()
	_, err := rc.Read(buf)
	if err == nil {
		t.Error("Should have panicked before checking error")
	}
}

func TestReadCloser_WithNilCloser(t *testing.T) {
	// Test with nil closer
	reader := &mockReader{data: "test"}
	rc := NewReadCloser(reader, nil)
	if rc == nil {
		t.Fatal("NewReadCloser returned nil with nil closer")
	}

	// Reading should work
	buf := make([]byte, 10)
	n, err := rc.Read(buf)
	if err != nil {
		t.Errorf("Read failed with nil closer: %v", err)
	}
	if string(buf[:n]) != "test" {
		t.Errorf("Expected 'test', got %q", string(buf[:n]))
	}

	// Closing should panic when closer is nil
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic when closing with nil closer")
		}
	}()
	err = rc.Close()
	// Should not reach here
	t.Error("Should have panicked before this point")
}

func TestReadCloser_WithStandardLibraryTypes(t *testing.T) {
	// Test with strings.Reader and a mock closer
	reader := strings.NewReader("hello from strings.Reader")
	closer := &mockCloser{}
	rc := NewReadCloser(reader, closer)

	// Read all data
	data, err := io.ReadAll(rc)
	if err != nil {
		t.Fatalf("ReadAll failed: %v", err)
	}
	if string(data) != "hello from strings.Reader" {
		t.Errorf("Expected 'hello from strings.Reader', got %q", string(data))
	}

	// Close should work
	err = rc.Close()
	if err != nil {
		t.Errorf("Close failed: %v", err)
	}
	if !closer.closed {
		t.Error("Closer was not called")
	}
}