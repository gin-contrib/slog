package slog

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

/*
Fn is a function type that takes a gin.Context and a *slog.Logger as parameters,
and returns a *slog.Logger. It is typically used to modify or enhance the logger
within the context of a Gin HTTP request.
*/
type Fn func(*gin.Context, *slog.Logger) *slog.Logger

/*
EventFn is a function type that takes a gin.Context and a *slog.Record as parameters,
and returns a *slog.Record. It is typically used to modify or enhance the record
within the context of a Gin HTTP request.
*/
type EventFn func(*gin.Context, *slog.Record) *slog.Record

/*
Skipper defines a function to skip middleware. It takes a gin.Context as input
and returns a boolean indicating whether to skip the middleware for the given context.
*/
type Skipper func(c *gin.Context) bool

// config holds logger middleware settings.
type config struct {
	logger                    Fn                    // custom logger function
	context                   EventFn               // gin.Context to log context
	utc                       bool                  // use UTC time
	skipPath                  []string              // exact path to skip
	skipPathRegexps           []*regexp.Regexp      // regex path to skip
	skip                      Skipper               // function to skip logging
	output                    io.Writer             // log output writer
	defaultLevel              slog.Level            // <400 log level
	clientErrorLevel          slog.Level            // 400-499 log level
	serverErrorLevel          slog.Level            // >=500 log level
	pathLevels                map[string]slog.Level // per-path <400 log level
	message                   string                // log message
	specificLevelByStatusCode map[int]slog.Level    // status-specific log level
	withRequestHeader         bool                  // log all headers
	hiddenRequestHeaders      map[string]struct{}   // hidden headers (lower-case)
}

const loggerKey = "_gin-contrib/logger_"

/*
SetLogger returns a gin.HandlerFunc (middleware) that logs requests using slog.
It accepts a variadic number of Option functions to customize the logger's behavior.

The logger configuration includes:
  - defaultLevel: the default logging level (default: slog.LevelInfo).
  - clientErrorLevel: the logging level for client errors (default: slog.LevelWarn).
  - serverErrorLevel: the logging level for server errors (default: slog.LevelError).
  - output: the output writer for the logger (default: gin.DefaultWriter).
  - skipPath: a list of paths to skip logging.
  - skipPathRegexps: a list of regular expressions to skip logging for matching paths.
  - logger: a custom logger function to use instead of the default logger.

The middleware logs the following request details:
  - method: the HTTP method of the request.
  - path: the URL path of the request.
  - ip: the client's IP address.
  - user_agent: the User-Agent header of the request.
  - status: the HTTP status code of the response.
  - latency: the time taken to process the request.
  - body_size: the size of the response body.

The logging level for each request is determined based on the response status code:
  - clientErrorLevel for 4xx status codes.
  - serverErrorLevel for 5xx status codes.
  - defaultLevel for other status codes.
  - Custom levels can be set for specific paths using the pathLevels configuration.
*/
func SetLogger(opts ...Option) gin.HandlerFunc {
	cfg := &config{
		defaultLevel:     slog.LevelInfo,
		clientErrorLevel: slog.LevelWarn,
		serverErrorLevel: slog.LevelError,
		output:           os.Stderr,
		message:          "Request",
		hiddenRequestHeaders: map[string]struct{}{
			"authorization": {},
			"cookie":        {},
			"set-cookie":    {},
			"x-auth-token":  {},
			"x-csrf-token":  {},
			"x-xsrf-token":  {},
			"user-agent":    {}, // Optional: Include user-agent in hidden headers
		},
	}

	// Apply each option to the config
	for _, o := range opts {
		o.apply(cfg)
	}

	// Create a set of paths to skip logging
	skip := map[string]struct{}{}
	for _, route := range cfg.skipPath {
		skip[route] = struct{}{}
	}

	// Initialize the base logger
	handler := slog.NewTextHandler(cfg.output, &slog.HandlerOptions{
		Level: cfg.defaultLevel,
	})
	l := slog.New(handler)

	return func(c *gin.Context) {
		rl := l
		if cfg.logger != nil {
			rl = cfg.logger(c, l)
		}

		start := time.Now()
		route := c.Request.URL.Path
		query := c.Request.URL.RawQuery
		c.Set(loggerKey, rl)

		c.Next()

		skipRoute := route
		if query != "" {
			skipRoute += "?" + query
		}

		track := !shouldSkipLogging(skipRoute, skip, cfg, c)
		if !track {
			return
		}

		end := time.Now()
		if cfg.utc {
			end = end.UTC()
		}

		msg := cfg.message
		if len(c.Errors) > 0 {
			msg += " with errors: " + c.Errors.String()
		}

		latency := end.Sub(start)
		status := c.Writer.Status()
		method := c.Request.Method
		userAgent := c.Request.UserAgent()
		ip := c.ClientIP()
		referer := c.Request.Referer()

		level := getLogLevel(cfg, c, route)
		record := slog.NewRecord(end, level, msg, 0)
		record.Add("status", status)
		record.Add("method", method)
		record.Add("path", route)
		record.Add("query", query)
		record.Add("route", c.FullPath())
		record.Add("ip", ip)
		record.Add("latency", latency)
		record.Add("referer", referer)
		record.Add("user_agent", userAgent)
		record.Add("body_size", c.Writer.Size())

		// Add each HTTP request header as a separate log field if enabled
		if cfg.withRequestHeader && c.Request.Header != nil {
			headers := make(map[string]any, len(c.Request.Header))
			for k, v := range c.Request.Header {
				keyLower := strings.ToLower(k)
				if _, hidden := cfg.hiddenRequestHeaders[keyLower]; hidden {
					continue
				}
				headers[k] = v
			}
			record.Add("headers", headers)
		}

		recPtr := &record
		if cfg.context != nil {
			recPtr = cfg.context(c, recPtr)
		}

		_ = rl.Handler().Handle(c.Request.Context(), *recPtr)
	}
}

/*
ParseLevel parses a string representation of a log level and returns the corresponding slog.Level.
It takes a single argument:
  - levelStr: a string representing the log level (e.g., "debug", "info", "warn", "error").

It returns:
  - slog.Level: the parsed log level.
  - error: an error if the log level string is invalid.
*/
func ParseLevel(levelStr string) (slog.Level, error) {
	switch levelStr {
	case "debug":
		return slog.LevelDebug, nil
	case "info":
		return slog.LevelInfo, nil
	case "warn", "warning":
		return slog.LevelWarn, nil
	case "error":
		return slog.LevelError, nil
	default:
		return slog.LevelInfo, fmt.Errorf("unknown slog level: %s", levelStr)
	}
}

func shouldSkipLogging(route string, skip map[string]struct{}, cfg *config, c *gin.Context) bool {
	if _, ok := skip[route]; ok || (cfg.skip != nil && cfg.skip(c)) {
		return true
	}
	for _, reg := range cfg.skipPathRegexps {
		if reg.MatchString(route) {
			return true
		}
	}
	return false
}

func getLogLevel(cfg *config, c *gin.Context, route string) slog.Level {
	if lvl, has := cfg.specificLevelByStatusCode[c.Writer.Status()]; has {
		return lvl
	}
	if c.Writer.Status() >= http.StatusBadRequest && c.Writer.Status() < http.StatusInternalServerError {
		return cfg.clientErrorLevel
	}
	if c.Writer.Status() >= http.StatusInternalServerError {
		return cfg.serverErrorLevel
	}
	if lvl, has := cfg.pathLevels[route]; has {
		return lvl
	}
	return cfg.defaultLevel
}

/*
Get retrieves the *slog.Logger instance from the given gin.Context.
It assumes that the logger has been previously set in the context with the key loggerKey.
If the logger is not found, it will panic.

Parameters:

	c - the gin.Context from which to retrieve the logger.

Returns:

	*slog.Logger - the logger instance stored in the context.
*/
func Get(c *gin.Context) *slog.Logger {
	return c.MustGet(loggerKey).(*slog.Logger)
}
