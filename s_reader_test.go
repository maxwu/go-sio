package go_sio

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"
)

func TestNewStreamReader(t *testing.T) {
	tests := []struct {
		name     string
		reader   io.Reader
		filter   StringLineFilter
		expected *StreamReader
	}{
		{
			name:     "nil reader",
			reader:   nil,
			filter:   NopFilter,
			expected: nil,
		},
		{
			name:     "valid reader with filter",
			reader:   strings.NewReader("test"),
			filter:   NopFilter,
			expected: &StreamReader{}, // Just check it's not nil
		},
		{
			name:     "valid reader with nil filter",
			reader:   strings.NewReader("test"),
			filter:   nil,
			expected: &StreamReader{}, // Should use NopFilter
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sr := NewStreamReader(tt.reader, tt.filter)
			if tt.expected == nil {
				if sr != nil {
					t.Error("Expected nil StreamReader")
				}
			} else {
				if sr == nil {
					t.Error("Expected non-nil StreamReader")
				}
			}
		})
	}
}

func TestStreamReader_Read_NilReceiver(t *testing.T) {
	var sr *StreamReader
	buf := make([]byte, 10)
	n, err := sr.Read(buf)
	if err != ErrNilReader {
		t.Errorf("Expected ErrNilReader, got %v", err)
	}
	if n != 0 {
		t.Errorf("Expected 0 bytes, got %d", n)
	}
}

func TestStreamReader_Read_BasicFunctionality(t *testing.T) {
	data := "line1\nline2\nline3\n"
	reader := strings.NewReader(data)
	sr := NewStreamReader(reader, NopFilter)

	// Read first line
	buf := make([]byte, 10)
	n, err := sr.Read(buf)
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}
	if string(buf[:n]) != "line1\n" {
		t.Errorf("Expected 'line1\\n', got %q", string(buf[:n]))
	}

	// Read second line
	buf = make([]byte, 10)
	n, err = sr.Read(buf)
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}
	if string(buf[:n]) != "line2\n" {
		t.Errorf("Expected 'line2\\n', got %q", string(buf[:n]))
	}

	// Read third line
	buf = make([]byte, 10)
	n, err = sr.Read(buf)
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}
	if string(buf[:n]) != "line3\n" {
		t.Errorf("Expected 'line3\\n', got %q", string(buf[:n]))
	}

	// Should return EOF on next read
	buf = make([]byte, 10)
	_, err = sr.Read(buf)
	if err != io.EOF {
		t.Errorf("Expected EOF, got %v", err)
	}
	if n != 0 {
		t.Errorf("Expected 0 bytes on EOF, got %d", n)
	}
}

func TestStreamReader_Read_WithFilter(t *testing.T) {
	data := "keep\nskip\nkeep\n"
	reader := strings.NewReader(data)
	
	// Filter that skips lines containing "skip"
	filter := func(line string) (string, error) {
		if strings.Contains(line, "skip") {
			return "", nil
		}
		return strings.ToUpper(line), nil
	}
	
	sr := NewStreamReader(reader, filter)

	// First read should get "KEEP\n"
	buf := make([]byte, 10)
	n, err := sr.Read(buf)
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}
	if string(buf[:n]) != "KEEP\n" {
		t.Errorf("Expected 'KEEP\\n', got %q", string(buf[:n]))
	}

	// Second read should skip "skip" line and get "KEEP\n"
	buf = make([]byte, 10)
	n, err = sr.Read(buf)
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}
	if string(buf[:n]) != "KEEP\n" {
		t.Errorf("Expected 'KEEP\\n', got %q", string(buf[:n]))
	}

	// Should return EOF on next read
	buf = make([]byte, 10)
	_, err = sr.Read(buf)
	if err != io.EOF {
		t.Errorf("Expected EOF, got %v", err)
	}
}

func TestStreamReader_Read_FilterError(t *testing.T) {
	data := "line1\nline2\n"
	reader := strings.NewReader(data)
	expectedErr := errors.New("filter error")
	
	filter := func(line string) (string, error) {
		if strings.Contains(line, "line2") {
			return "", expectedErr
		}
		return line, nil
	}
	
	sr := NewStreamReader(reader, filter)

	// First read should succeed
	buf := make([]byte, 10)
	n, err := sr.Read(buf)
	if err != nil {
		t.Fatalf("First read failed: %v", err)
	}
	if string(buf[:n]) != "line1\n" {
		t.Errorf("Expected 'line1\\n', got %q", string(buf[:n]))
	}

	// Second read should return filter error
	buf = make([]byte, 10)
	n, err = sr.Read(buf)
	if err != expectedErr {
		t.Errorf("Expected filter error %v, got %v", expectedErr, err)
	}
	if n != 0 {
		t.Errorf("Expected 0 bytes on error, got %d", n)
	}
}

func TestStreamReader_Read_EmptyLines(t *testing.T) {
	data := "line1\n\nline3\n"
	reader := strings.NewReader(data)
	sr := NewStreamReader(reader, NopFilter)

	// Read first line
	buf := make([]byte, 10)
	n, err := sr.Read(buf)
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}
	if string(buf[:n]) != "line1\n" {
		t.Errorf("Expected 'line1\\n', got %q", string(buf[:n]))
	}

	// Read empty line
	buf = make([]byte, 10)
	n, err = sr.Read(buf)
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}
	if string(buf[:n]) != "\n" {
		t.Errorf("Expected '\\n', got %q", string(buf[:n]))
	}

	// Read third line
	buf = make([]byte, 10)
	n, err = sr.Read(buf)
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}
	if string(buf[:n]) != "line3\n" {
		t.Errorf("Expected 'line3\\n', got %q", string(buf[:n]))
	}
}

func TestStreamReader_Read_NoFinalNewline(t *testing.T) {
	data := "line1\nline2"
	reader := strings.NewReader(data)
	sr := NewStreamReader(reader, NopFilter)

	// Read first line
	buf := make([]byte, 10)
	n, err := sr.Read(buf)
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}
	if string(buf[:n]) != "line1\n" {
		t.Errorf("Expected 'line1\\n', got %q", string(buf[:n]))
	}

	// Read second line (no trailing newline)
	buf = make([]byte, 10)
	n, err = sr.Read(buf)
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}
	if string(buf[:n]) != "line2" {
		t.Errorf("Expected 'line2', got %q", string(buf[:n]))
	}

	// Should return EOF on next read
	buf = make([]byte, 10)
	n, err = sr.Read(buf)
	if err != io.EOF {
		t.Errorf("Expected EOF, got %v", err)
	}
}

func TestSplit(t *testing.T) {
	tests := []struct {
		name      string
		data      []byte
		atEOF     bool
		advance   int
		token     []byte
		err       error
	}{
		{
			name:    "empty data at EOF",
			data:    []byte{},
			atEOF:   true,
			advance: 0,
			token:   nil,
			err:     nil,
		},
		{
			name:    "single line with newline",
			data:    []byte("hello\n"),
			atEOF:   false,
			advance: 6,
			token:   []byte("hello\n"),
			err:     nil,
		},
		{
			name:    "line without newline at EOF",
			data:    []byte("hello"),
			atEOF:   true,
			advance: 5,
			token:   []byte("hello"),
			err:     nil,
		},
		{
			name:    "line without newline not at EOF",
			data:    []byte("hello"),
			atEOF:   false,
			advance: 0,
			token:   nil,
			err:     nil,
		},
		{
			name:    "multiple lines",
			data:    []byte("line1\nline2\nline3"),
			atEOF:   false,
			advance: 6,
			token:   []byte("line1\n"),
			err:     nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			advance, token, err := split(tt.data, tt.atEOF)
			if advance != tt.advance {
				t.Errorf("Expected advance %d, got %d", tt.advance, advance)
			}
			if !bytes.Equal(token, tt.token) {
				t.Errorf("Expected token %q, got %q", tt.token, token)
			}
			if err != tt.err {
				t.Errorf("Expected error %v, got %v", tt.err, err)
			}
		})
	}
}

func TestNopFilter(t *testing.T) {
	input := "test string"
	output, err := NopFilter(input)
	if err != nil {
		t.Errorf("NopFilter returned error: %v", err)
	}
	if output != input {
		t.Errorf("NopFilter changed input: expected %q, got %q", input, output)
	}
}

func TestNewJSONFilterReadCloser(t *testing.T) {
	// Test data with valid and invalid JSON
	data := `{"valid": "json"}
invalid line
{"another": "valid"}
not json at all
["array", "is", "valid"]
`
	reader := io.NopCloser(strings.NewReader(data))
	
	rc := NewJSONFilterReadCloser(reader)
	if rc == nil {
		t.Fatal("NewJSONFilterReadCloser returned nil")
	}

	// Read all data
	result, err := io.ReadAll(rc)
	if err != nil {
		t.Fatalf("ReadAll failed: %v", err)
	}

	// Should only contain valid JSON lines
	expected := `{"valid": "json"}
{"another": "valid"}
["array", "is", "valid"]
`
	if string(result) != expected {
		t.Errorf("Expected:\n%s\nGot:\n%s", expected, string(result))
	}

	// Test closing
	err = rc.Close()
	if err != nil {
		t.Errorf("Close failed: %v", err)
	}
}

func TestNewJSONFilterReadCloser_EmptyInput(t *testing.T) {
	reader := io.NopCloser(strings.NewReader(""))
	rc := NewJSONFilterReadCloser(reader)
	
	result, err := io.ReadAll(rc)
	if err != nil {
		t.Fatalf("ReadAll failed: %v", err)
	}
	
	if len(result) != 0 {
		t.Errorf("Expected empty result, got %q", string(result))
	}
}

func TestNewJSONFilterReadCloser_OnlyInvalidJSON(t *testing.T) {
	data := `not json
also not json
{invalid json}
`
	reader := io.NopCloser(strings.NewReader(data))
	rc := NewJSONFilterReadCloser(reader)
	
	result, err := io.ReadAll(rc)
	if err != nil {
		t.Fatalf("ReadAll failed: %v", err)
	}
	
	if len(result) != 0 {
		t.Errorf("Expected empty result, got %q", string(result))
	}
}