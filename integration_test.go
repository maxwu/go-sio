package go_sio

import (
	"bytes"
	"io"
	"strings"
	"testing"
)

// TestIntegration_RealWorldUsage demonstrates real-world usage scenarios
func TestIntegration_RealWorldUsage(t *testing.T) {
	t.Run("StreamReader with JSON filter", func(t *testing.T) {
		// Simulate log data with mixed JSON and non-JSON lines
		logData := `2024-01-01 INFO: Starting application
{"timestamp": "2024-01-01T10:00:00Z", "level": "info", "message": "User logged in", "user_id": 123}
2024-01-01 DEBUG: Processing request
{"timestamp": "2024-01-01T10:01:00Z", "level": "error", "message": "Database connection failed", "error": "timeout"}
Not valid JSON
{"timestamp": "2024-01-01T10:02:00Z", "level": "warn", "message": "High memory usage", "memory_percent": 85.5}
`

		reader := io.NopCloser(strings.NewReader(logData))
		jsonReader := NewJSONFilterReadCloser(reader)

		result, err := io.ReadAll(jsonReader)
		if err != nil {
			t.Fatalf("Failed to read JSON lines: %v", err)
		}

		expected := `{"timestamp": "2024-01-01T10:00:00Z", "level": "info", "message": "User logged in", "user_id": 123}
{"timestamp": "2024-01-01T10:01:00Z", "level": "error", "message": "Database connection failed", "error": "timeout"}
{"timestamp": "2024-01-01T10:02:00Z", "level": "warn", "message": "High memory usage", "memory_percent": 85.5}
`

		if string(result) != expected {
			t.Errorf("JSON filter didn't work as expected.\nExpected:\n%s\nGot:\n%s", expected, string(result))
		}

		err = jsonReader.Close()
		if err != nil {
			t.Errorf("Failed to close JSON reader: %v", err)
		}
	})

	t.Run("TeeReader for monitoring", func(t *testing.T) {
		// Simulate monitoring a data stream while processing it
		data := "line1\nline2\nline3\nline4\n"
		source := io.NopCloser(strings.NewReader(data))
		
		// Buffer to capture the data for monitoring
		var monitor bytes.Buffer
		
		teeReader := NewTeeReaderCloser(source, &monitor)

		// Process the data (simulate reading and processing)
		var processed bytes.Buffer
		_, err := io.Copy(&processed, teeReader)
		if err != nil {
			t.Fatalf("Failed to process data: %v", err)
		}

		// Verify the data was processed correctly
		if processed.String() != data {
			t.Errorf("Data processing failed. Expected %q, got %q", data, processed.String())
		}

		// Verify the monitor captured the same data
		if monitor.String() != data {
			t.Errorf("Monitoring failed. Expected %q, got %q", data, monitor.String())
		}

		err = teeReader.Close()
		if err != nil {
			t.Errorf("Failed to close tee reader: %v", err)
		}
	})

	t.Run("Custom filter for line processing", func(t *testing.T) {
		// Simulate processing log lines with custom formatting
		logData := "INFO: Starting service\nERROR: Connection failed\nDEBUG: Processing data\nWARN: Low disk space\n"
		
		// Filter that only passes ERROR and WARN lines and converts to uppercase
		errorWarnFilter := func(line string) (string, error) {
			if strings.Contains(line, "ERROR:") || strings.Contains(line, "WARN:") {
				return strings.ToUpper(line), nil
			}
			return "", nil // Skip other lines
		}

		reader := strings.NewReader(logData)
		streamReader := NewStreamReader(reader, errorWarnFilter)

		result, err := io.ReadAll(streamReader)
		if err != nil {
			t.Fatalf("Failed to read filtered lines: %v", err)
		}

		expected := "ERROR: CONNECTION FAILED\nWARN: LOW DISK SPACE\n"
		if string(result) != expected {
			t.Errorf("Custom filter didn't work as expected.\nExpected: %q\nGot: %q", expected, string(result))
		}
	})

	t.Run("Chained readers", func(t *testing.T) {
		// Chain multiple readers together
		data := `{"level": "info", "msg": "test1"}
not json
{"level": "error", "msg": "test2"}
also not json
{"level": "debug", "msg": "test3"}
`

		// First, filter only JSON lines
		source := io.NopCloser(strings.NewReader(data))
		jsonReader := NewJSONFilterReadCloser(source)

		// Then, monitor the JSON stream
		var monitor bytes.Buffer
		teeReader := NewTeeReaderCloser(jsonReader, &monitor)

		// Finally, apply a custom filter to only keep error logs
		errorFilter := func(line string) (string, error) {
			if strings.Contains(line, `"level": "error"`) {
				return line, nil
			}
			return "", nil
		}
		
		errorReader := NewStreamReader(teeReader, errorFilter)

		result, err := io.ReadAll(errorReader)
		if err != nil {
			t.Fatalf("Failed to read chained result: %v", err)
		}

		// Should only contain the error JSON line
		expected := `{"level": "error", "msg": "test2"}
`
		if string(result) != expected {
			t.Errorf("Chained readers didn't work as expected.\nExpected: %q\nGot: %q", expected, string(result))
		}

		// Monitor should contain all JSON lines that passed through
		expectedMonitor := `{"level": "info", "msg": "test1"}
{"level": "error", "msg": "test2"}
{"level": "debug", "msg": "test3"}
`
		if monitor.String() != expectedMonitor {
			t.Errorf("Monitor didn't capture all JSON lines.\nExpected: %q\nGot: %q", expectedMonitor, monitor.String())
		}

		err = teeReader.Close()
		if err != nil {
			t.Errorf("Failed to close tee reader: %v", err)
		}
	})
}

// BenchmarkStreamReader exercises StreamReader in common filter scenarios.
func BenchmarkStreamReader(b *testing.B) {
	passThroughData := strings.Repeat("test line\n", 1000)
	dropData := strings.Repeat("keep\nskip\n", 500)           // drops half the lines
	transformData := strings.Repeat("lowercase line\n", 1000) // uppercases all lines

	dropFilter := func(s string) (string, error) {
		if strings.HasPrefix(s, "keep") {
			return s, nil
		}
		return "", nil
	}
	transformFilter := func(s string) (string, error) {
		return strings.ToUpper(s), nil
	}

	benchmarks := []struct {
		name   string
		data   string
		filter StringLineFilter
	}{
		{"NoFilter", passThroughData, NopFilter},
		{"FilterDropHalf", dropData, dropFilter},
		{"FilterTransform", transformData, transformFilter},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				reader := strings.NewReader(bm.data)
				sr := NewStreamReader(reader, bm.filter)
				_, _ = io.ReadAll(sr)
			}
		})
	}
}

// BenchmarkJSONFilter measures filtering of mixed JSON/non-JSON lines.
func BenchmarkJSONFilter(b *testing.B) {
	data := strings.Repeat(`{"test": "data"}
not json
`, 500)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		reader := io.NopCloser(strings.NewReader(data))
		jr := NewJSONFilterReadCloser(reader)
		_, _ = io.ReadAll(jr)
		_ = jr.Close()
	}
}

// BenchmarkTeeReader measures teeing a stream into a side buffer.
func BenchmarkTeeReader(b *testing.B) {
	data := strings.Repeat("benchmark data\n", 1000)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		reader := io.NopCloser(strings.NewReader(data))
		var writer bytes.Buffer
		tr := NewTeeReaderCloser(reader, &writer)
		_, _ = io.ReadAll(tr)
		_ = tr.Close()
	}
}
