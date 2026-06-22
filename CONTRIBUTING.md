# Contributing to go-sio

## Requirements

- Go 1.26 or newer
- golangci-lint 2.12
- No third-party production or test dependencies
- 100% unit-test coverage for production code

## Workflow

1. Make a focused, reviewable change.
2. Format every changed Go file with `gofmt`.
3. Add or update tests for every behavior change.
4. Run the validation commands below.
5. Open a pull request that explains the change and its test coverage.

```bash
golangci-lint run ./...
go test -count=1 -race ./... -coverprofile=coverage.out
go tool cover -func=coverage.out | grep -qE '^total:.*100\.0%$'
go test -bench . -run '^$' ./...
```

All checks must pass before a pull request is ready for review.
