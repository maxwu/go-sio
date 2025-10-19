# go-sio

Minimal helpers for streaming line-based IO in Go with no external dependencies.

This small library provides utilities for reading streamed data line-by-line with configurable filtering, wrapping readers with closers, and a tee-style reader that captures the stream while still exposing an io.ReadCloser. The package has no extra dependencies beyond the Go standard library.

## Key components

- **StreamReader**: an io.Reader that reads input line-by-line and applies a configurable filter function to each line. Useful for processing logs or other newline-delimited streams without loading the whole stream into memory.
- **NewJSONFilterReadCloser**: wraps an existing io.ReadCloser and only yields lines that are valid JSON.
- **NewTeeReaderCloser**: a combination of io.TeeReader and an io.Closer — useful when you want to copy the stream to another writer while preserving Close.
- **NewReadCloser**: create a simple io.ReadCloser from an io.Reader and an io.Closer.

## Installation

This project uses Go modules. From your module, add the dependency with:

```bash
go get github.com/go-sio@latest
```

Import in your code:

```go
import "github.com/go-sio"
```

## Quick examples

### 1. StreamReader — read a stream line-by-line and filter lines

```go
package main

import (
    "fmt"
    "io"
    "strings"
    "github.com/go-sio"
)

func main() {
    data := "one\ntwo\nthree\n"
    r := strings.NewReader(data)
    f := func(s string) (string, error) {
        if s == "" {
            return "", nil
        }
        return strings.ToUpper(s), nil
    }
    sr := go_sio.NewStreamReader(r, f)
    if sr == nil {
        panic("failed to create StreamReader")
    }
    out, _ := io.ReadAll(sr)
    fmt.Print(string(out))
}
```

### 2. NewJSONFilterReadCloser — only emit lines that are valid JSON

```go
package main

import (
    "fmt"
    "io"
    "os"
    "github.com/go-sio"
)

func main() {
    f, _ := os.Open("stream.log")
    defer f.Close()
    rc := go_sio.NewJSONFilterReadCloser(f)
    defer rc.Close()
    b, _ := io.ReadAll(rc)
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
    "os"
    "github.com/go-sio"
)

func main() {
    f, _ := os.Open("stream.log")
    defer f.Close()
    var buf bytes.Buffer
    trc := go_sio.NewTeeReaderCloser(f, &buf)
    defer trc.Close()
    _, _ = io.Copy(os.Stdout, trc)
    fmt.Println("Captured:", buf.String())
}
```

## API reference (summary)

 `type StringLineFilter func(string) (string, error)`
    - Filter applied to each line read by StreamReader. Return empty string to drop the line, return a non-empty string to emit it. Return an error to abort reading.
    - Create a new StreamReader. If `r` is nil this returns nil. If `f` is nil a no-op filter is used.
    - Wrap a ReadCloser and only yield lines that are valid JSON (uses encoding/json.Valid).
    - Wraps the provided ReadCloser with an io.TeeReader that writes to `w` while still exposing Close.
    - Utility to combine a Reader and a Closer into a single io.ReadCloser.

- `func NewTeeReaderCloser(r io.ReadCloser, w io.Writer) *TeeReaderCloser`
  - Wraps the provided ReadCloser with an io.TeeReader that writes to `w` while still exposing Close.
- `type ReadCloser struct { io.Reader; io.Closer }`
- `func NewReadCloser(r io.Reader, c io.Closer) *ReadCloser`
  - Utility to combine a Reader and a Closer into a single io.ReadCloser.

## Notes and behaviour

- StreamReader reads using a bufio.Scanner with a custom split function that keeps newline characters in emitted tokens. The provided filter receives the full line (including newline) as a string. Returning the empty string drops that line from output.
- NewStreamReader will return nil when passed a nil reader — callers should check for this.
- StreamReader's Read returns ErrNilReader (from the package) if the receiver is nil.

## Contributing

Contributions are welcome. Please open issues or pull requests for bugs, feature requests, or improvements.

## License

This project is MIT licensed — see the `LICENSE` file.
