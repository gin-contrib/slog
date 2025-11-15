# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**gin-slog** is a Gin middleware library that provides structured logging using Go's standard `log/slog` package (Go 1.21+). It allows customizable log formatting, level control per path/status, and context injection for HTTP requests. The middleware logs request details (method, path, status, latency, etc.) with configurable output and filtering.

**Key requirements:**

- Go 1.24+
- Depends on `github.com/gin-gonic/gin` v1.11.0+

## Development Commands

### Testing

```bash
# Run all tests
go test -v -covermode=atomic -coverprofile=coverage.out

# Run specific test
go test -run ^TestName$

# Run tests with coverage
go test -v -covermode=atomic -coverprofile=coverage.out
```

### Linting and Formatting

```bash
# Run golangci-lint (uses .golangci.yml config)
golangci-lint run --verbose

# Format code (required before committing)
go fmt ./...

# Run gofumpt (stricter formatter, part of golangci-lint)
gofumpt -l -w .

# Run goimports
goimports -w .

# Vet code
go vet ./...
```

### Building

```bash
# Build all packages
go build ./...

# Run example
go run _example/main.go
```

## Architecture

The codebase is a single-package library with a clean separation of concerns:

### Core Components

**slog.go** - Main middleware implementation:

- `SetLogger(opts ...Option)` returns a `gin.HandlerFunc` that logs HTTP requests
- `Get(c *gin.Context) *slog.Logger` retrieves the logger from Gin context
- Internal `config` struct holds all middleware settings
- Log level determination: checks specific status codes first, then 4xx/5xx ranges, then path-specific levels, finally default level
- Headers filtering: sensitive headers (authorization, cookie, etc.) are hidden by default when request header logging is enabled

**options.go** - Configuration via functional options pattern:

- All options implement the `Option` interface with `apply(*config)` method
- Options: `WithLogger`, `WithContext`, `WithUTC`, `WithSkipPath`, `WithSkipPathRegexps`, `WithWriter`, `WithDefaultLevel`, `WithClientErrorLevel`, `WithServerErrorLevel`, `WithPathLevel`, `WithMessage`, `WithSpecificLogLevelByStatusCode`, `WithRequestHeader`, `WithHiddenRequestHeaders`

### Function Types

- `Fn func(*gin.Context, *slog.Logger) *slog.Logger` - custom logger injection
- `EventFn func(*gin.Context, *slog.Record) *slog.Record` - log record modification
- `Skipper func(c *gin.Context) bool` - conditional logging skip

### Logger Storage

The middleware stores the logger in Gin's context with key `loggerKey = "_gin-contrib/logger_"`. Access it via `slog.Get(c)` in handlers.

## Code Style

This project follows Go standard conventions with strict linting:

- **Import order:** Standard library → third-party → local (separate blocks)
- **Naming:** Exported: `CamelCase`, private: `camelCase`
- **Error handling:** Return errors; avoid panics except for programmer errors
- **Formatting:** Use `gofmt` and `gofumpt` (extra-rules enabled) - **required before commits**
- **No sensitive data:** Never log authorization headers, cookies, tokens (hidden by default)
- **Function design:** Small, focused, minimal side effects

### Linters Enabled

See `.golangci.yml` for full list. Key linters: `bodyclose`, `errcheck`, `gosec`, `govet`, `ineffassign`, `misspell`, `staticcheck`, `unused`, `whitespace`

## Testing CI

Tests run on Go 1.23, 1.24, 1.25 on Ubuntu. Coverage uploaded to Codecov.

## Important Implementation Details

1. **Log level priority:** Status-specific code > 4xx/5xx range > path-specific > default
2. **Header filtering:** When `WithRequestHeader(true)`, headers are logged except those in `hiddenRequestHeaders` map (case-insensitive)
3. **Skip logic:** Path checked against exact strings first, then regex patterns, then custom `Skipper` function
4. **Timestamp:** End time is captured after request completion; use `WithUTC(true)` for UTC timestamps
5. **Error messages:** If `c.Errors` contains entries, they're appended to the log message
