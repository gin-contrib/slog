package main

import (
	"net/http"

	"github.com/gin-contrib/slog"
	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.New()

	// Add slog middleware with default settings
	r.Use(slog.SetLogger(
		slog.WithRequestHeader(true),
	))

	// Example route
	r.GET("/", func(c *gin.Context) {
		slog.Get(c).Info("Hello World!")
		c.String(http.StatusOK, "ok")
	})

	r.GET("/ok", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "ok",
		})
	})

	// Example route to handle requests with headers
	r.GET("/header", func(c *gin.Context) {
		// Get specific headers
		userAgent := c.GetHeader("User-Agent")
		authorization := c.GetHeader("Authorization")
		customHeader := c.GetHeader("X-Custom-Header")

		// Log header information
		logger := slog.Get(c)
		logger.Info("Received request with headers",
			"user_agent", userAgent,
			"authorization", authorization,
			"custom_header", customHeader,
		)

		c.JSON(http.StatusOK, gin.H{
			"message": "Headers received",
			"headers": gin.H{
				"user_agent":    userAgent,
				"authorization": authorization,
				"custom_header": customHeader,
			},
		})
	})

	r.Run()
}
