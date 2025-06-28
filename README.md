# gin-slog

[![Run Tests](https://github.com/gin-contrib/slog/actions/workflows/go.yml/badge.svg?branch=main)](https://github.com/gin-contrib/slog/actions/workflows/go.yml)
[![codecov](https://codecov.io/gh/gin-contrib/slog/branch/main/graph/badge.svg)](https://codecov.io/gh/gin-contrib/slog)
[![Go Report Card](https://goreportcard.com/badge/github.com/gin-contrib/slog)](https://goreportcard.com/report/github.com/gin-contrib/slog)
[![GoDoc](https://pkg.go.dev/badge/github.com/gin-contrib/slog?utm_source=godoc)](https://pkg.go.dev/github.com/gin-contrib/slog)

Gin middleware for Go 1.23+ [`slog`](https://pkg.go.dev/log/slog) logging.

## Overview

**gin-slog** is a Gin middleware that provides structured logging using Go's standard [`slog`](https://pkg.go.dev/log/slog) package (available since Go 1.21). It allows you to customize log formatting, target, level per path/status, and injection of additional context for each request, making it easy to produce standardized, production-ready logs in your Gin applications.

## Features

- Log HTTP requests with structured output via `slog`
- Configurable log levels per status code or endpoint path
- Easily skip logging for specific paths (by string or regexp)
- Supports custom log writers and messages
- Add custom fields or alter records/context per request
- Output in standard text format compatible with `slog`
- Fully extensible via Options pattern

## Installation

```sh
go get github.com/gin-contrib/slog
```

Requires Go 1.23+.

## Usage

### Basic Example

```go
package main

import (
  "net/http"

  "github.com/gin-contrib/slog"
  "github.com/gin-gonic/gin"
)

func main() {
  r := gin.New()

  // Add slog middleware with default settings
  r.Use(slog.SetLogger())

  // Example route
  r.GET("/", func(c *gin.Context) {
    slog.Get(c).Info("Hello World!")
    c.String(http.StatusOK, "ok")
  })

  r.Run()
}
```

### Customization

You can customize the middleware via options:

```go
import (
  "os"
  "log/slog"
  "regexp"
  
  "github.com/gin-gonic/gin"
  "github.com/gin-contrib/slog"
)

func main() {
  r := gin.New()

  r.Use(slog.SetLogger(
    // Change log writer
    slog.WithWriter(os.Stdout),
    // Use UTC timestamps
    slog.WithUTC(true),
    // Skip health check and static routes
    slog.WithSkipPath([]string{"/healthz", "/metrics"}),
    // Skip paths matching pattern (e.g., static assets)
    slog.WithSkipPathRegexps(regexp.MustCompile(`\.ico$`), regexp.MustCompile(`^/static/`)),
    // Change default log levels
    slog.WithDefaultLevel(slog.LevelDebug),
    slog.WithClientErrorLevel(slog.LevelWarn),
    slog.WithServerErrorLevel(slog.LevelError),
    // Log message customization
    slog.WithMessage("Handled request"),
    // Set specific log level for a given path
    slog.WithPathLevel(map[string]slog.Level{"/foo": slog.LevelInfo}),
    // Set log level by status code
    slog.WithSpecificLogLevelByStatusCode(map[int]slog.Level{418: slog.LevelDebug}),
    // Inject custom info/context into logs
    slog.WithContext(func(c *gin.Context, rec *slog.Record) *slog.Record {
      rec.Add("user_agent", c.Request.UserAgent())
      return rec
    }),
    // Provide your own logger (to add global fields, etc.)
    slog.WithLogger(func(c *gin.Context, l *slog.Logger) *slog.Logger {
      return l.With("request_id", c.GetString("request_id"))
    }),
    // Custom Skipper (function: skip logging if ...), example:
    slog.WithSkipper(func(c *gin.Context) bool {
      return c.Request.Method == "OPTIONS"
    }),
  ))

  r.Run()
}
```

## Logged Fields

Each HTTP request log will include by default:

- `status` (int): Response HTTP status code
- `method` (string): HTTP method
- `path` (string): URL path
- `ip` (string): Client IP address
- `latency` (duration): Time to handle request
- `user_agent` (string): Client's User-Agent header
- `body_size` (int): Size of the response body

Additional fields can be injected via `WithContext`.

Log level is determined by status code, per-path configuration, or explicit mapping (see below).

## API

### Middleware

#### `slog.SetLogger(opts ...Option) gin.HandlerFunc`

Creates a Gin middleware handler. All customization is done via options (see next section).

#### `slog.Get(c *gin.Context) *slog.Logger`

Retrieves the underlying `*slog.Logger` from Gin's context. Access this in your handlers for structured custom logging.

---

### Options

All the options below can be passed to `SetLogger()`.

| Option                                                 | Description                                                                              |
| ------------------------------------------------------ | ---------------------------------------------------------------------------------------- |
| `WithLogger(fn)`                                       | Inject a custom logger for each request: `func(*gin.Context, *slog.Logger) *slog.Logger` |
| `WithContext(fn)`                                      | Alter the log record per request: `func(*gin.Context, *slog.Record) *slog.Record`        |
| `WithWriter(w io.Writer)`                              | Set log output (default: `gin.DefaultWriter`; e.g., `os.Stdout`)                         |
| `WithMessage(msg string)`                              | Set a custom message for each log line (default: `"Request"`)                            |
| `WithSkipPath([]string)`                               | List of URL paths to skip logging                                                        |
| `WithSkipPathRegexps(...*regexp.Regexp)`               | Regexps to match paths to skip logging                                                   |
| `WithSkipper(fn)`                                      | Custom Skipper function: `func(c *gin.Context) bool`—return `true` to skip this request  |
| `WithUTC(bool)`                                        | Use UTC instead of local time                                                            |
| `WithDefaultLevel(slog.Level)`                         | Level for requests with status < 400 (default: `Info`)                                   |
| `WithClientErrorLevel(slog.Level)`                     | Level for 4xx (default: `Warn`)                                                          |
| `WithServerErrorLevel(slog.Level)`                     | Level for 5xx (default: `Error`)                                                         |
| `WithPathLevel(map[string]slog.Level)`                 | Map of URL paths to log levels                                                           |
| `WithSpecificLogLevelByStatusCode(map[int]slog.Level)` | Set log level for specific status codes                                                  |

#### Parsing Levels

Use `slog.ParseLevel(str)` to convert strings like `"debug"`, `"info"`, `"warn"`, `"error"` to `slog.Level` values.

---

## License

[MIT License](LICENSE)
