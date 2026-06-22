# go-sio

[![codecov](https://codecov.io/gh/maxwu/go-sio/graph/badge.svg?token=VG2FF2QYUI)](https://codecov.io/gh/maxwu/go-sio)

Minimal helpers for streaming line-based IO in Go with no external dependencies.

This small library provides utilities for reading streamed data line-by-line with configurable filtering, wrapping readers with closers, and a tee-style reader that captures the stream while still exposing an io.ReadCloser. The package has no extra dependencies beyond the Go standard library.

## Key components

- **StreamReader**: an io.Reader that reads input line-by-line and applies a configurable filter function to each line. Useful for processing logs or other newline-delimited streams incrementally.
- **NewJSONFilterReadCloser**: wraps an existing io.ReadCloser and only yields lines that are valid JSON.
- **NewTeeReaderCloser**: a combination of io.TeeReader and an io.Closer — useful when you want to copy the stream to another writer while preserving Close.
- **NewReadCloser**: create a simple io.ReadCloser from an io.Reader and an io.Closer.

## Installation

This project uses Go modules and requires Go 1.26 or newer. From your module, add the dependency with:

```bash
go get github.com/maxwu/go-sio@latest
```

Import in your code:

```go
import "github.com/maxwu/go-sio"

// You can also give the import an explicit package name:
// import go_sio "github.com/maxwu/go-sio"
// (Go automatically treats the package name as go_sio even without the alias.)
```

## Quick examples

### 1. StreamReader — read a stream line-by-line and filter lines

```go
package main

import (
    "fmt"
    "io"
    "log"
    "strings"

    "github.com/maxwu/go-sio"
)

func main() {
    data := "one\n\ntwo\nthree"
    r := strings.NewReader(data)
    f := func(s string) (string, error) {
        if s == "\n" {
            return "", nil
        }
        return strings.ToUpper(s), nil
    }
    sr := go_sio.NewStreamReader(r, f)
    if sr == nil {
        panic("failed to create StreamReader")
    }
    out, err := io.ReadAll(sr)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Print(string(out))
}
```

### 2. NewJSONFilterReadCloser — only emit lines that are valid JSON

```go
package main

import (
    "fmt"
    "io"
    "log"
    "os"

    "github.com/maxwu/go-sio"
)

func main() {
    f, err := os.Open("stream.log")
    if err != nil {
        log.Fatal(err)
    }
    rc := go_sio.NewJSONFilterReadCloser(f)

    b, readErr := io.ReadAll(rc)
    closeErr := rc.Close()
    if readErr != nil {
        log.Fatal(readErr)
    }
    if closeErr != nil {
        log.Fatal(closeErr)
    }
    fmt.Print(string(b))
}
```

### 3. NewTeeReaderCloser — capture the stream while still returning an io.ReadCloser

```go
package main

import (
    "bytes"
    "fmt"
    "io"
    "log"
    "os"

    "github.com/maxwu/go-sio"
)

func main() {
    f, err := os.Open("stream.log")
    if err != nil {
        log.Fatal(err)
    }
    var buf bytes.Buffer
    trc := go_sio.NewTeeReaderCloser(f, &buf)

    _, copyErr := io.Copy(os.Stdout, trc)
    closeErr := trc.Close()
    if copyErr != nil {
        log.Fatal(copyErr)
    }
    if closeErr != nil {
        log.Fatal(closeErr)
    }
    fmt.Println("Captured:", buf.String())
}
```

## API reference (summary)

- `type StringLineFilter func(string) (string, error)`: filter applied to each line read by StreamReader. Return an empty string to drop a line; return an error to abort reading.
- `var ErrNilReader error`: returned when calling StreamReader.Read on a nil receiver.
- `var NopFilter StringLineFilter`: a pass-through filter used when `nil` is provided.
- `type StreamReader`: an io.Reader that emits filtered lines.
- `func NewStreamReader(r io.Reader, f StringLineFilter) *StreamReader`: creates a StreamReader; returns nil when `r` is nil; falls back to NopFilter when `f` is nil.
- `func NewJSONFilterReadCloser(r io.ReadCloser) io.ReadCloser`: wraps `r` and only yields lines that are valid JSON (uses `encoding/json.Valid`).
- `type TeeReaderCloser struct { ... }`
- `func NewTeeReaderCloser(r io.ReadCloser, w io.Writer) *TeeReaderCloser`: wraps `r` with an io.TeeReader that writes to `w` while preserving `Close`.
- `type ReadCloser struct { io.Reader; io.Closer }`
- `func NewReadCloser(r io.Reader, c io.Closer) *ReadCloser`: utility to combine a Reader and a Closer into a single io.ReadCloser.

## Notes and behaviour

- StreamReader reads using a bufio.Scanner with a custom split function. The filter receives the newline terminator when one is present; the final line may be passed without a newline. Returning the empty string drops that line from output.
- The scanner uses Go's default maximum token size of approximately 64 KiB. Reading a longer line returns a `bufio.Scanner: token too long` error.
- NewJSONFilterReadCloser accepts any complete JSON value recognized by `encoding/json.Valid`, including objects, arrays, strings, numbers, booleans, and null.
- Closing a ReadCloser returned by NewJSONFilterReadCloser or NewTeeReaderCloser closes the original reader. Callers should close only the wrapper.
- NewStreamReader will return nil when passed a nil reader — callers should check for this.
- StreamReader's Read returns ErrNilReader (from the package) if the receiver is nil.

## Development

Development requires Go 1.26 and golangci-lint 2.12. The library must remain free of third-party dependencies, and all production code must retain 100% unit-test coverage.

Format changed Go files with `gofmt`, then run the same checks used by CI:

```bash
golangci-lint run ./...
go test -count=1 -race ./... -coverprofile=coverage.out
go tool cover -func=coverage.out | grep -qE '^total:.*100\.0%$'
go test -bench . -run '^$' ./...
```

See [CONTRIBUTING.md](./CONTRIBUTING.md) for the contributor workflow and [AGENTS.md](./AGENTS.md) for coding-agent constraints.

## Contributing

Contributions are welcome. Please read [CONTRIBUTING.md](./CONTRIBUTING.md) before opening a pull request.

## License

This project is MIT licensed — see [LICENSE](./LICENSE).
