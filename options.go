package slog

import (
	"io"
	"log/slog"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
)

type Option interface {
	apply(*config)
}

var _ Option = (*optionFunc)(nil)

type optionFunc func(*config)

func (o optionFunc) apply(c *config) {
	o(c)
}

// WithLogger sets a logger function to the config.
func WithLogger(fn func(*gin.Context, *slog.Logger) *slog.Logger) Option {
	return optionFunc(func(c *config) {
		c.logger = fn
	})
}

// WithSkipPathRegexps appends regexp rules for skipPathRegexps in config.
func WithSkipPathRegexps(regs ...*regexp.Regexp) Option {
	return optionFunc(func(c *config) {
		if len(regs) == 0 {
			return
		}

		c.skipPathRegexps = append(c.skipPathRegexps, regs...)
	})
}

// WithUTC sets UTC mode for logging time.
func WithUTC(s bool) Option {
	return optionFunc(func(c *config) {
		c.utc = s
	})
}

// WithSkipPath sets URL path list for skipPath in config.
func WithSkipPath(s []string) Option {
	return optionFunc(func(c *config) {
		c.skipPath = s
	})
}

// WithPathLevel sets path-specific logging levels.
func WithPathLevel(m map[string]slog.Level) Option {
	return optionFunc(func(c *config) {
		c.pathLevels = m
	})
}

// WithWriter sets the log output destination.
func WithWriter(s io.Writer) Option {
	return optionFunc(func(c *config) {
		c.output = s
	})
}

// WithDefaultLevel sets config defaultLevel (<400 status).
func WithDefaultLevel(lvl slog.Level) Option {
	return optionFunc(func(c *config) {
		c.defaultLevel = lvl
	})
}

// WithClientErrorLevel sets client error log level (400-499).
func WithClientErrorLevel(lvl slog.Level) Option {
	return optionFunc(func(c *config) {
		c.clientErrorLevel = lvl
	})
}

// WithServerErrorLevel sets server error log level.
func WithServerErrorLevel(lvl slog.Level) Option {
	return optionFunc(func(c *config) {
		c.serverErrorLevel = lvl
	})
}

// WithSkipper sets a function to skip logging for certain requests.
func WithSkipper(s Skipper) Option {
	return optionFunc(func(c *config) {
		c.skip = s
	})
}

// WithContext sets a custom context handler for slog.Record.
func WithContext(fn func(*gin.Context, *slog.Record) *slog.Record) Option {
	return optionFunc(func(c *config) {
		c.context = fn
	})
}

// WithMessage sets a custom log message for requests.
func WithMessage(message string) Option {
	return optionFunc(func(c *config) {
		c.message = message
	})
}

// WithSpecificLogLevelByStatusCode sets specific log level per HTTP status.
func WithSpecificLogLevelByStatusCode(statusCodes map[int]slog.Level) Option {
	return optionFunc(func(c *config) {
		c.specificLevelByStatusCode = statusCodes
	})
}

// WithRequestHeader enables/disables logging all HTTP request headers.
func WithRequestHeader(enabled bool) Option {
	return optionFunc(func(c *config) {
		c.withRequestHeader = enabled
	})
}

// WithHiddenRequestHeaders sets request header names to be hidden. Only works with WithRequestHeader enabled.
func WithHiddenRequestHeaders(headers []string) Option {
	return optionFunc(func(c *config) {
		c.hiddenRequestHeaders = make(map[string]struct{}, len(headers))
		for _, h := range headers {
			c.hiddenRequestHeaders[strings.ToLower(h)] = struct{}{}
		}
	})
}
